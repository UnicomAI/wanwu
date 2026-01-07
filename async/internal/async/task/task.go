package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/UnicomAI/wanwu/async/internal/db/model"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

// Task 运行中的任务代理 goroutine safe
type Task struct {
	taskID  uint32
	initCtx string

	task async_task.ITask
	log  async_config.Logger

	mutex   sync.Mutex
	stop    chan struct{}
	stopped bool
}

func NewTask(dbTask *model.AsyncTask, asyncTask async_task.ITask, log async_config.Logger) *Task {
	return &Task{
		taskID:  dbTask.ID,
		initCtx: string(dbTask.Ctx),
		task:    asyncTask,
		log:     log,
		stop:    make(chan struct{}, 1),
	}
}

func (t *Task) ID() uint32 {
	return t.taskID
}

func (t *Task) InitCtx() string {
	return t.initCtx
}

func (t *Task) Running(ctx context.Context) (<-chan async_task.IReport, error) {
	if t.CheckStop() {
		return nil, fmt.Errorf("async task %v already send stop event %v", t.taskID, t.stopped)
	}
	return t.task.Running(ctx, t.initCtx, t.stop), nil
}

func (t *Task) Deleting(ctx context.Context) (<-chan async_task.IReport, error) {
	if t.CheckStop() {
		return nil, fmt.Errorf("async task %v already send stop", t.taskID)
	}
	return t.task.Deleting(ctx, t.initCtx, t.stop), nil
}

func (t *Task) CheckStop() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.stopped
}

func (t *Task) SendStop() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.stopped {
		return fmt.Errorf("async task %v already send stop", t.taskID)
	}
	t.stopped = true
	t.stop <- struct{}{}
	t.log.Debugf("async task %v send stop", t.taskID)
	return nil
}
