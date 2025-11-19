package lock

import (
	"context"
	"time"
)

type DistributedLockConfig struct {
	Timeout     time.Duration // 锁超时时间 [EN] lock timeout
	SyncAcquire bool          // 同步获取锁 [EN] Acquire locks synchronously
}

type DistributedLockService interface {
	AcquireLock(ctx context.Context, lockKey string, config *DistributedLockConfig) error
	ReleaseLock(ctx context.Context, lockKey string) error
}
