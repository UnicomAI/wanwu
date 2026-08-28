package async_task

import (
	"context"

	async "github.com/gromitlee/go-async"
	"github.com/gromitlee/go-async/pkg/async/async_component/pending"
	"gorm.io/gorm"
)

// InitAsync 初始化 go-async 运行时并启动已注册业务任务的 pending 队列。
// 由 app-service main 在 DB 初始化完成后调用。
func InitAsync(ctx context.Context, db *gorm.DB) error {
	pendingRun, err := pending.NewPendingRun(db, nil)
	if err != nil {
		return err
	}
	options := []async.AsyncOption{
		async.WithPendingRunQueue(pendingRun),
	}
	if err = async.Init(ctx, db, options...); err != nil {
		return err
	}
	if err = InitAllService(); err != nil {
		return err
	}
	return nil
}

// StopAsync 停止 go-async 运行时。
func StopAsync() {
	async.Stop()
}
