package rsautil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	// rsaKeyBits RSA密钥位数
	rsaKeyBits = 2048

	// privateKeyFileMode 私钥文件权限（仅所有者可读写）
	privateKeyFileMode = 0600

	// publicKeyFileMode 公钥文件权限（所有者读写，其他只读）
	publicKeyFileMode = 0644
)

// KeyPair 表示一个RSA密钥对
type KeyPair struct {
	KeyID      string
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// KeyManager RSA密钥管理器
type KeyManager struct {
	currentKey *KeyPair
	keyCache   map[string]*KeyPair
	mu         sync.RWMutex
	config     Config
}

// Config RSA密钥配置
type Config struct {
	// 私钥文件路径
	PrivateKeyPath string `json:"private_key_path" mapstructure:"private_key_path"`
	// 公钥文件路径
	PublicKeyPath string `json:"public_key_path" mapstructure:"public_key_path"`
	// 密钥轮换周期（天），默认90天
	KeyRotationDays int `json:"key_rotation_days" mapstructure:"key_rotation_days"`
}

var (
	manager *KeyManager
	once    sync.Once
)

// InitKeyManager 初始化密钥管理器（单例）
func InitKeyManager(config Config) error {
	var initErr error
	once.Do(func() {
		manager = &KeyManager{
			keyCache: make(map[string]*KeyPair),
			config:   config,
		}
		initErr = func() error {
			manager.mu.Lock()
			defer manager.mu.Unlock()
			return manager.loadFromFile()
		}()
	})
	return initErr
}

// GetManager 获取密钥管理器实例
func GetManager() *KeyManager {
	if manager == nil {
		log.Panicf("rsa key manager not initialized, please call InitKeyManager first")
	}
	return manager
}

// loadFromFile 从文件加载密钥对，如果文件不存在则自动生成
// 调用方需持有 km.mu 写锁
func (km *KeyManager) loadFromFile() error {
	if km.config.PrivateKeyPath == "" {
		return errors.New("private key path is empty")
	}

	// 检查私钥文件是否存在，不存在则自动生成
	if _, err := os.Stat(km.config.PrivateKeyPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat private key file failed: %w", err)
		}
		// 文件不存在，自动生成密钥对
		if genErr := generateKeyFilesAtomic(km.config.PrivateKeyPath, km.config.PublicKeyPath); genErr != nil {
			return fmt.Errorf("auto-generate RSA key pair failed: %w", genErr)
		}
	}

	// 读取私钥文件
	privateKeyData, err := os.ReadFile(km.config.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("read private key file failed: %w", err)
	}

	// 解析 PEM 格式
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return errors.New("failed to decode PEM block containing private key")
	}

	// 解析 PKCS8 格式私钥
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse private key failed: %w", err)
	}
	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return errors.New("not an RSA private key")
	}

	// 生成密钥ID
	keyID := generateKeyID()

	// 设置过期时间
	rotationDays := km.config.KeyRotationDays
	if rotationDays <= 0 {
		rotationDays = 90 // 默认90天
	}

	keyPair := &KeyPair{
		KeyID:      keyID,
		PublicKey:  &privateKey.PublicKey,
		PrivateKey: privateKey,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().AddDate(0, 0, rotationDays),
	}

	km.currentKey = keyPair
	km.keyCache[keyID] = keyPair

	//log.Infof("RSA key loaded from file, keyID: %s, expires at: %s", keyID, keyPair.ExpiresAt.Format("2006-01-02"))
	return nil
}

// GetPublicKey 获取当前公钥信息（供前端使用）
func (km *KeyManager) GetPublicKey() (keyID string, publicKeyPEM string, expiresIn int64, err error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.currentKey == nil {
		return "", "", 0, errors.New("no available RSA key")
	}

	// 序列化公钥为 PEM 格式
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(km.currentKey.PublicKey)
	if err != nil {
		return "", "", 0, fmt.Errorf("marshal public key failed: %w", err)
	}

	publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	// 建议前端缓存5分钟
	expiresIn = 300

	return km.currentKey.KeyID, publicKeyPEM, expiresIn, nil
}

// Decrypt 使用指定密钥解密数据
func (km *KeyManager) Decrypt(keyID string, ciphertextBase64 string) ([]byte, error) {
	km.mu.RLock()
	keyPair, exists := km.keyCache[keyID]
	km.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid key id: %s", keyID)
	}

	// Base64 解码
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	// RSA-OAEP 解密
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, keyPair.PrivateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA decrypt failed: %w", err)
	}

	return plaintext, nil
}

// DecryptWithCurrentKey 使用当前密钥解密（兼容无keyID的情况）
func (km *KeyManager) DecryptWithCurrentKey(ciphertextBase64 string) ([]byte, error) {
	km.mu.RLock()
	if km.currentKey == nil {
		km.mu.RUnlock()
		return nil, errors.New("no available RSA key")
	}
	keyID := km.currentKey.KeyID
	km.mu.RUnlock()

	return km.Decrypt(keyID, ciphertextBase64)
}

// generateKeyID 生成密钥ID
func generateKeyID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		// 使用 crypto/rand 生成随机数
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 如果随机数生成失败，使用时间戳作为后备
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
			continue
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// RotateKey 手动轮换密钥（用于运维操作）
func (km *KeyManager) RotateKey() error {
	km.mu.Lock()
	defer km.mu.Unlock()

	// 从文件重新加载密钥
	return km.loadFromFile()
}

// EnsureKeyFiles 确保RSA密钥文件存在，如果私钥文件不存在则自动生成
// 这是一个独立的函数，可在KeyManager之外使用（如OAuth RSA密钥初始化）
func EnsureKeyFiles(privateKeyPath, publicKeyPath string) error {
	if _, err := os.Stat(privateKeyPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat private key file failed: %w", err)
		}
		// 文件不存在，自动生成密钥对
		return generateKeyFilesAtomic(privateKeyPath, publicKeyPath)
	}
	return nil
}

// GenerateKeyFiles 生成RSA密钥对并保存到指定路径（供运维工具调用）
// 如果文件已存在则返回错误，避免覆盖
// 使用原子写入（先写临时文件再 rename）防止半写入状态
func GenerateKeyFiles(privateKeyPath, publicKeyPath string) error {
	// 检查私钥文件是否已存在
	if _, err := os.Stat(privateKeyPath); err == nil {
		return fmt.Errorf("private key file already exists: %s", privateKeyPath)
	}

	return generateKeyFilesAtomic(privateKeyPath, publicKeyPath)
}

// generateKeyFilesAtomic 原子写入密钥文件
// 先写入同目录下的临时文件，完成后 rename 到目标路径，防止半写入状态
// 使用 O_EXCL 排他创建锁文件防止多实例并发生成同一密钥文件
func generateKeyFilesAtomic(privateKeyPath, publicKeyPath string) error {
	// 文件锁，防止多实例并发生成
	lockPath := privateKeyPath + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			// 锁文件已存在，说明其他实例正在生成，等待其完成
			//log.Infof("RSA key generation in progress by another instance, waiting...")
			if waitErr := waitForFile(privateKeyPath, 30*time.Second); waitErr != nil {
				return fmt.Errorf("timed out waiting for RSA key generation by another instance: %w", waitErr)
			}
			return nil
		}
		return fmt.Errorf("create lock file failed: %w", err)
	}
	// 获取锁成功，确保退出时清理
	defer func() {
		if err := lockFile.Close(); err != nil {
			log.Warnf("close lock file failed: %v", err)
		}
		if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
			log.Warnf("remove lock file failed: %v", err)
		}
	}()

	// 获取锁后再次检查，防止竞态
	if _, err := os.Stat(privateKeyPath); err == nil {
		//log.Infof("RSA private key file already exists (generated by another instance): %s", privateKeyPath)
		return nil
	}

	// 生成 RSA 密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeyBits)
	if err != nil {
		return fmt.Errorf("generate RSA key pair failed: %w", err)
	}

	// 编码私钥（PKCS8 PEM）
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("marshal private key failed: %w", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 原子写入私钥文件
	if err := atomicWriteFile(privateKeyPath, privateKeyPEM, privateKeyFileMode); err != nil {
		return fmt.Errorf("write private key file failed: %w", err)
	}

	// 原子写入公钥文件
	if publicKeyPath != "" {
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return fmt.Errorf("marshal public key failed: %w", err)
		}
		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		})
		if err := atomicWriteFile(publicKeyPath, publicKeyPEM, publicKeyFileMode); err != nil {
			return fmt.Errorf("write public key file failed: %w", err)
		}
	}

	//log.Infof("RSA key pair generated: private_key=%s, public_key=%s", privateKeyPath, publicKeyPath)
	return nil
}

// atomicWriteFile 原子写入文件：先写入同目录临时文件，再 rename 到目标路径
// 如果 rename 失败（如 Docker bind mount 不支持跨文件系统 rename），回退到直接写入
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory failed: %w", err)
	}

	// 写入临时文件
	tmpFile, err := os.CreateTemp(dir, ".rsa-key-tmp-")
	if err != nil {
		// 如果无法创建临时文件，直接写入目标文件
		return os.WriteFile(path, data, perm)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		if closeErr := tmpFile.Close(); closeErr != nil {
			log.Warnf("close temp file failed: %v", closeErr)
		}
		if removeErr := os.Remove(tmpPath); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Warnf("remove temp file failed: %v", removeErr)
		}
		return fmt.Errorf("write temp file failed: %w", err)
	}

	if err := tmpFile.Chmod(perm); err != nil {
		if closeErr := tmpFile.Close(); closeErr != nil {
			log.Warnf("close temp file failed: %v", closeErr)
		}
		if removeErr := os.Remove(tmpPath); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Warnf("remove temp file failed: %v", removeErr)
		}
		return fmt.Errorf("chmod temp file failed: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Warnf("remove temp file failed: %v", removeErr)
		}
		return fmt.Errorf("close temp file failed: %w", err)
	}

	// 原子 rename 到目标路径
	if err := os.Rename(tmpPath, path); err != nil {
		// rename 失败（Docker bind mount 等场景），回退到直接写入
		if removeErr := os.Remove(tmpPath); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Warnf("remove temp file failed: %v", removeErr)
		}
		//log.Infof("rename failed for %s (likely Docker bind mount), falling back to direct write: %v", path, err)
		if err := os.WriteFile(path, data, perm); err != nil {
			return fmt.Errorf("direct write file failed: %w", err)
		}
	}

	return nil
}

// waitForFile 等待文件出现且非空，直到超时
func waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if info, err := os.Stat(path); err == nil && info.Size() > 0 {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("file not found or empty within %s: %s", timeout, path)
}
