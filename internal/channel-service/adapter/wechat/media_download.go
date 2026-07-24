package wechat

import (
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// 本文件实现微信入站媒体（文件/图片）的 CDN 下载与 AES-128-ECB 解密，
// 对齐 OpenClaw 微信插件 src/cdn/pic-decrypt.ts 的 downloadAndDecryptBuffer。
// 出站（SendFile）的上传加密仍用 adapter.go 的 aesECBEncrypt，二者对称。

// parseAesKey 把 CDNMedia.aes_key（base64）解析为 16 字节 AES 密钥。
// 兼容两种编码（对齐 OpenClaw pic-decrypt.ts:40 parseAesKey）：
//   - base64(原始 16 字节)：图片（image_item.aeskey 转 base64 后）
//   - base64(hex 32 字符)：文件/语音/视频，需先 base64 解出 hex 串再 hex 解析
func parseAesKey(aesKeyBase64 string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(aesKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("aes_key base64 decode failed: %w", err)
	}
	if len(decoded) == 16 {
		return decoded, nil
	}
	if len(decoded) == 32 {
		// base64 解出的是 32 字符 hex 串，再 hex 解析成 16 字节
		if raw, decErr := hex.DecodeString(string(decoded)); decErr == nil && len(raw) == 16 {
			return raw, nil
		}
	}
	return nil, fmt.Errorf("aes_key must decode to 16 raw bytes or 32-char hex string, got %d bytes", len(decoded))
}

// aesECBDecrypt AES-128-ECB 解密并去 PKCS7 padding（对称于 adapter.go 的 aesECBEncrypt）。
func aesECBDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	if len(ciphertext) == 0 || len(ciphertext)%bs != 0 {
		return nil, fmt.Errorf("ciphertext length %d not multiple of block size %d", len(ciphertext), bs)
	}
	plaintext := make([]byte, len(ciphertext))
	for start := 0; start < len(ciphertext); start += bs {
		block.Decrypt(plaintext[start:start+bs], ciphertext[start:start+bs])
	}
	// 去 PKCS7 padding
	pad := int(plaintext[len(plaintext)-1])
	if pad == 0 || pad > bs || pad > len(plaintext) {
		return nil, fmt.Errorf("invalid pkcs7 padding: %d", pad)
	}
	return plaintext[:len(plaintext)-pad], nil
}

// downloadAndDecrypt 下载并 AES-128-ECB 解密 CDN 媒体文件，返回明文字节。
// URL 取值：优先 media.FullURL，否则 cdnBaseURL + "/download?encrypted_query_param=" + url.QueryEscape(encQP)。
// 复用传入 httpClient（无 Bearer，鉴权内嵌在 URL 签名里）。
func downloadAndDecrypt(ctx context.Context, httpClient *http.Client, media *CDNMedia, aesKeyBase64, cdnBaseURL string) ([]byte, error) {
	downloadURL := media.FullURL
	if downloadURL == "" {
		if media.EncryptQueryParam == "" {
			return nil, fmt.Errorf("no download url: both full_url and encrypt_query_param empty")
		}
		downloadURL = fmt.Sprintf("%s/download?encrypted_query_param=%s", cdnBaseURL, url.QueryEscape(media.EncryptQueryParam))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build cdn download request failed: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cdn download network error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cdn download http %d: %s", resp.StatusCode, truncateWechat(string(body)))
	}
	ciphertext, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read cdn ciphertext failed: %w", err)
	}
	key, err := parseAesKey(aesKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("parse aes key failed: %w", err)
	}
	plaintext, err := aesECBDecrypt(key, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("aes decrypt failed: %w", err)
	}
	return plaintext, nil
}

// extMimeOverrides 按扩展名内置 MIME 映射（参考 OpenClaw media/mime.ts）。
// 标准库 mime.TypeByExtension 对 docx/xlsx/pptx 等容器格式常返回空（取决于系统 mime.types），
// 故内置兜底，避免 docx/xlsx 被当 application/octet-stream 影响 WGA 文件解析。
var extMimeOverrides = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".txt":  "text/plain",
	".csv":  "text/csv",
	".md":   "text/markdown",
	".html": "text/html",
	".htm":  "text/html",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".webp": "image/webp",
	".wav":  "audio/wav",
	".mp3":  "audio/mpeg",
	".mp4":  "video/mp4",
}

// mimeTypeByExt 按文件名扩展名推断 MIME 类型，推断失败回退 application/octet-stream。
// 不用 http.DetectContentType 嗅探（会把 docx/xlsx zip 容器误判为 octet-stream）。
func mimeTypeByExt(fileName string) string {
	ext := strings.ToLower(path.Ext(fileName))
	if m, ok := extMimeOverrides[ext]; ok {
		return m
	}
	if mt := mime.TypeByExtension(ext); mt != "" {
		return mt
	}
	return "application/octet-stream"
}
