package async

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/UnicomAI/wanwu/async/internal/async"
	"github.com/UnicomAI/wanwu/async/internal/async/config"
	"github.com/UnicomAI/wanwu/async/internal/db/model"
	"github.com/UnicomAI/wanwu/async/internal/db/trans"
	"github.com/UnicomAI/wanwu/async/internal/tools"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_component"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_component/pending"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

var _mgr *async.Mgr

func Init(ctx context.Context, db *gorm.DB, options ...AsyncOption) error {
	if _mgr != nil {
		return ErrMgrAlreadyInit
	}
	var err error

	cfg := config.Config{
		Log:               tools.DefaultLog(),
		RunMaxConcurrency: 5,
		RunTaskInterval:   1,
	}
	for _, opt := range options {
		if cfg, err = opt.apply(cfg); err != nil {
			return err
		}
	}
	// pendingRun, pendingDel
	if cfg.PendingRun == nil {
		cfg.PendingRun = pending.NewPendingRunDefault(db, cfg.Log)
	}
	if cfg.PendingDel == nil {
		cfg.PendingDel = pending.NewPendingDelDefault(db, cfg.Log)
	}

	if _mgr, err = async.NewMgr(db, cfg); err != nil {
		return err
	}
	return _mgr.Run(ctx)
}

func Stop() {
	if _mgr == nil {
		return
	}
	_mgr.Stop()
	_mgr = nil
}

func RegisterTask(taskTyp uint32, newTask async_task.ITaskFunc) error {
	if _mgr == nil {
		return ErrMgrNotInit
	}
	return _mgr.RegisterTask(taskTyp, newTask)
}

func CreateTask(ctx context.Context, user, group string, taskTyp uint32, taskCtx string, autoRun bool) (uint32, error) {
	if _mgr == nil {
		return 0, ErrMgrNotInit
	}
	return _mgr.CreateTask(ctx, user, group, taskTyp, taskCtx, autoRun)
}

func ChangeTaskGroup(ctx context.Context, taskID uint32, group string) error {
	if _mgr == nil {
		return ErrMgrNotInit
	}
	return _mgr.ChangeTaskGroup(ctx, taskID, group)
}

func RunTask(ctx context.Context, taskID uint32) error {
	if _mgr == nil {
		return ErrMgrNotInit
	}
	return _mgr.UserRun(ctx, taskID)
}

func DeleteTask(ctx context.Context, taskID uint32) error {
	if _mgr == nil {
		return ErrMgrNotInit
	}
	return _mgr.UserDelete(ctx, taskID)
}

func PauseTask(ctx context.Context, taskID uint32) error {
	if _mgr == nil {
		return ErrMgrNotInit
	}
	return _mgr.UserPause(ctx, taskID)
}

func GetTask(ctx context.Context, taskID uint32) (*async_task.Task, error) {
	if _mgr == nil {
		return nil, ErrMgrNotInit
	}
	dbTask, err := _mgr.GetTask(ctx, taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return convert(dbTask)
}

func GetTasks(ctx context.Context, user, group string, taskTypes []uint32, states []async_task.State, offset, limit int32) ([]*async_task.Task, error) {
	if _mgr == nil {
		return nil, ErrMgrNotInit
	}
	if len(states) == 0 {
		states = append(states,
			async_task.StateInit,
			async_task.StatePending,
			async_task.StateRunning,
			async_task.StateCanceling,
			async_task.StatePause,
			async_task.StateFinished,
			async_task.StateFailed)
	}
	var status []trans.TaskStatus
	for _, state := range states {
		switch state {
		case async_task.StateInit:
			status = append(status, trans.TaskStatus{S: trans.TaskStateInit, M: trans.TaskMarkNone})
		case async_task.StatePending:
			status = append(status, trans.TaskStatus{S: trans.TaskStatePending, M: trans.TaskMarkRun})
		case async_task.StateRunning:
			status = append(status,
				trans.TaskStatus{S: trans.TaskStateRunning, M: trans.TaskMarkRun},
				trans.TaskStatus{S: trans.TaskStatePause, M: trans.TaskMarkRun})
		case async_task.StateCanceling:
			status = append(status,
				trans.TaskStatus{S: trans.TaskStateRunning, M: trans.TaskMarkDelete},
				trans.TaskStatus{S: trans.TaskStateRunning, M: trans.TaskMarkPause})
		case async_task.StatePause:
			status = append(status, trans.TaskStatus{S: trans.TaskStatePause, M: trans.TaskMarkPause})
		case async_task.StateFinished:
			status = append(status, trans.TaskStatus{S: trans.TaskStateFinished, M: trans.TaskMarkRun})
		case async_task.StateFailed:
			status = append(status, trans.TaskStatus{S: trans.TaskStateFailed, M: trans.TaskMarkRun})
		default:
			return nil, fmt.Errorf("invalid state (%v)", state)
		}
	}
	dbTasks, err := _mgr.GetTasks(ctx, user, group, taskTypes, status, offset, limit)
	if err != nil {
		return nil, err
	}
	var tasks []*async_task.Task
	for _, dbTask := range dbTasks {
		if t, err := convert(dbTask); err == nil {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

// --- AsyncOption ---

func WithLogger(logger async_config.Logger) AsyncOption {
	return asyncOptionFunc(func(cfg config.Config) (config.Config, error) {
		if logger != nil {
			cfg.Log = logger
		} else {
			cfg.Log = tools.EmptyLog()
		}
		return cfg, nil
	})
}

func WithRunMaxConcurrency(max int) AsyncOption {
	return asyncOptionFunc(func(cfg config.Config) (config.Config, error) {
		if max <= 0 {
			return cfg, errors.New("invalid run max concurrency")
		}
		cfg.RunMaxConcurrency = max
		return cfg, nil
	})
}

func WithRunTaskIntervalSecond(interval int) AsyncOption {
	return asyncOptionFunc(func(cfg config.Config) (config.Config, error) {
		if interval <= 0 {
			return cfg, errors.New("invalid run task interval")
		}
		cfg.RunTaskInterval = interval
		return cfg, nil
	})
}

func WithPendingRunQueue(pendingRun async_component.IQueue) AsyncOption {
	return asyncOptionFunc(func(cfg config.Config) (config.Config, error) {
		if pendingRun != nil {
			cfg.PendingRun = pendingRun
		}
		return cfg, nil
	})
}

func WithPendingDelQueue(pendingDel async_component.IQueue) AsyncOption {
	return asyncOptionFunc(func(cfg config.Config) (config.Config, error) {
		if pendingDel != nil {
			cfg.PendingDel = pendingDel
		}
		return cfg, nil
	})
}

type AsyncOption interface {
	apply(cfg config.Config) (config.Config, error)
}

type asyncOptionFunc func(cfg config.Config) (config.Config, error)

func (fn asyncOptionFunc) apply(cfg config.Config) (config.Config, error) {
	return fn(cfg)
}

func convert(dbTask *model.AsyncTask) (*async_task.Task, error) {
	if dbTask == nil {
		return nil, ErrTaskNotFound
	}
	var state async_task.State
	status := trans.TaskStatus{S: dbTask.State, M: dbTask.Mark}
	switch status {
	// 任务创建初始状态，不会排队、运行
	case trans.TaskStatus{S: trans.TaskStateInit, M: trans.TaskMarkNone}:
		state = async_task.StateInit
	// 任务在pending.runQueue中排队，用户可见为排队中
	case trans.TaskStatus{S: trans.TaskStatePending, M: trans.TaskMarkRun}:
		state = async_task.StatePending
	// 任务在内存中执行run，用户可见为运行中
	case trans.TaskStatus{S: trans.TaskStateRunning, M: trans.TaskMarkRun}:
		state = async_task.StateRunning
	// 任务在内存中执行run，但用户标记删除，用户可见为取消中
	case trans.TaskStatus{S: trans.TaskStateRunning, M: trans.TaskMarkDelete}:
		state = async_task.StateCanceling
	// 任务在内存中执行run，用户标记暂停，用户可见为取消中
	case trans.TaskStatus{S: trans.TaskStateRunning, M: trans.TaskMarkPause}:
		state = async_task.StateCanceling
	// 任务暂停run，但用户可见为运行中
	case trans.TaskStatus{S: trans.TaskStatePause, M: trans.TaskMarkRun}:
		state = async_task.StateRunning
	// 任务暂停run，用户可见为暂停中
	case trans.TaskStatus{S: trans.TaskStatePause, M: trans.TaskMarkPause}:
		state = async_task.StatePause
	// 任务执行run结束，用户可见为结束
	case trans.TaskStatus{S: trans.TaskStateFinished, M: trans.TaskMarkRun}:
		state = async_task.StateFinished
	// 任务执行run失败，用户可见为失败
	case trans.TaskStatus{S: trans.TaskStateFailed, M: trans.TaskMarkRun}:
		state = async_task.StateFailed
	default:
		return nil, ErrTaskNotFound
	}
	return &async_task.Task{
		ID:        dbTask.ID,
		User:      dbTask.User,
		Group:     dbTask.Group,
		Type:      dbTask.Type,
		State:     state,
		CreatedAt: dbTask.CreatedAt,
		DoneAt:    dbTask.DoneAt,
		Ctx:       string(dbTask.Ctx),
	}, nil
}

var (
	ErrMgrNotInit     = errors.New("async mgr not init")
	ErrMgrAlreadyInit = errors.New("async mgr already init")

	ErrTaskNotFound = errors.New("async task not found")
)
