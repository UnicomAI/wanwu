package rsautil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
)

const (
	// challengeRedisKeyPrefix Challenge在Redis中的key前缀
	challengeRedisKeyPrefix = "wanwu:rsa:challenge:"
	// challengeTTL Challenge有效期（5分钟）
	challengeTTL = 5 * time.Minute
)

// challengeManager 全局Challenge管理器实例
var challengeManager *ChallengeManager

// ChallengeManager 基于Redis的Challenge管理器
// 负责Challenge的生成、存储和一次性消费校验
type ChallengeManager struct{}

// InitChallengeManager 初始化Challenge管理器
func InitChallengeManager() {
	challengeManager = &ChallengeManager{}
}

// GetChallengeManager 获取Challenge管理器实例
func GetChallengeManager() *ChallengeManager {
	if challengeManager == nil {
		panic("challenge manager not initialized, please call InitChallengeManager first")
	}
	return challengeManager
}

// GenerateChallenge 生成Challenge并存入Redis，返回challenge字符串
// Challenge由服务端生成，前端获取公钥时一并下发，嵌入cipher加密载荷中提交
func (m *ChallengeManager) GenerateChallenge(ctx context.Context) (string, error) {
	cli := redis.OP()
	if cli == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	// 生成32字节随机Challenge
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate challenge failed: %w", err)
	}
	challenge := hex.EncodeToString(b)

	// 存入Redis，value用于计数，TTL 5分钟
	key := challengeRedisKeyPrefix + challenge
	if _, err := cli.SetEx(ctx, key, "1", challengeTTL); err != nil {
		return "", fmt.Errorf("store challenge failed: %w", err)
	}

	return challenge, nil
}

// Validate 校验Challenge是否有效（仅检查，不消费）
// 返回true表示Challenge有效，false表示Challenge不存在（已过期或已使用）
func (m *ChallengeManager) Validate(ctx context.Context, challenge string) (bool, error) {
	cli := redis.OP()
	if cli == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	if challenge == "" {
		return false, fmt.Errorf("challenge is empty")
	}

	key := challengeRedisKeyPrefix + challenge
	val, err := cli.Get(ctx, key)
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return false, nil // key不存在（已过期或已使用）
		}
		return false, fmt.Errorf("redis get challenge failed: %w", err)
	}
	return val != "", nil
}

// ValidateAndConsume 校验Challenge是否有效，有效则一次性消费（删除）
// 返回true表示Challenge有效且已消费，false表示Challenge不存在（已过期或已使用）
func (m *ChallengeManager) ValidateAndConsume(ctx context.Context, challenge string) (bool, error) {
	cli := redis.OP()
	if cli == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	if challenge == "" {
		return false, fmt.Errorf("challenge is empty")
	}

	key := challengeRedisKeyPrefix + challenge
	// 使用Lua脚本保证原子性：检查存在 + 删除
	script := `
		local val = redis.call('GET', KEYS[1])
		if val then
			redis.call('DEL', KEYS[1])
			return 1
		end
		return 0
	`
	result, err := cli.Eval(ctx, script, []string{key})
	if err != nil {
		return false, fmt.Errorf("validate challenge failed: %w", err)
	}

	// Eval返回的是interface{}，需要转换为int64
	count, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected redis eval result type: %T", result)
	}

	return count == 1, nil
}
