package lock

import (
	"context"
	"time"
)

type DistributedLockConfig struct {
	Timeout     time.Duration // lock timeout
	SyncAcquire bool          // Acquire locks synchronously
}

type DistributedLockService interface {
	AcquireLock(ctx context.Context, lockKey string, config *DistributedLockConfig) error
	ReleaseLock(ctx context.Context, lockKey string) error
}
