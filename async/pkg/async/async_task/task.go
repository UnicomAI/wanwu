package async_task

import (
	"context"
)

// RunPhase 任务运行阶段
type RunPhase int32

const (
	RunPhaseFailed   RunPhase = -1 // 失败
	RunPhaseNormal   RunPhase = 1  // 正常
	RunPhaseFinished RunPhase = 2  // 完成
)

// IReport 任务上报interface，注意goroutine safe
type IReport interface {
	// Phase 上报运行阶段，bool表示finished/failed是否删除任务记录
	Phase() (RunPhase, bool)
	// Context 上报上下文
	Context() string
}

// ITask 异步任务interface
type ITask interface {
	// Running 任务执行，任务接收到stop消息后，须主动Report，之后退出即可
	Running(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan IReport
	// Deleting 删除任务
	Deleting(ctx context.Context, taskCtx string, stop <-chan struct{}) <-chan IReport
}

// ITaskFunc 异步任务初始化方法
type ITaskFunc func() ITask

// State 任务状态
type State int32

const (
	StateInit      State = 0 // 已创建
	StatePending   State = 1 // 排队中
	StateRunning   State = 2 // 运行中
	StateCanceling State = 3 // 取消中
	StatePause     State = 4 // 暂停
	StateFinished  State = 5 // 结束
	StateFailed    State = 6 // 失败
)

type Task struct {
	ID        uint32
	User      string
	Group     string
	Type      uint32
	State     State
	CreatedAt int64
	DoneAt    int64
	Ctx       string
}
