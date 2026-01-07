package async

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"

	"github.com/UnicomAI/wanwu/async/internal/async/config"
	"github.com/UnicomAI/wanwu/async/internal/async/deleting"
	"github.com/UnicomAI/wanwu/async/internal/async/fixing"
	"github.com/UnicomAI/wanwu/async/internal/async/running"
	"github.com/UnicomAI/wanwu/async/internal/async/task"
	"github.com/UnicomAI/wanwu/async/internal/db/client"
	"github.com/UnicomAI/wanwu/async/internal/db/model"
	"github.com/UnicomAI/wanwu/async/internal/db/trans"
	"github.com/UnicomAI/wanwu/async/internal/tools"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

// Mgr goroutine safe
type Mgr struct {
	taskTypes sync.Map // 任务注册 taskType -> newTask func

	log    async_config.Logger
	client client.ITaskClient

	runningMod  running.IModule
	deletingMod deleting.IModule
	cleaner     fixing.IClean

	mutex   sync.Mutex
	stopped bool
	stop    chan struct{}

	wg sync.WaitGroup
}

func NewMgr(db *gorm.DB, cfg config.Config) (*Mgr, error) {
	c, err := client.NewClient(db, cfg)
	if err != nil {
		return nil, err
	}
	return &Mgr{
		log:    cfg.Log,
		client: c,

		runningMod:  running.NewModule(c, cfg),
		deletingMod: deleting.NewModule(c, cfg),
		cleaner:     fixing.NewClean(c, cfg),

		stop: make(chan struct{}, 1),
	}, nil
}

func (m *Mgr) Run(ctx context.Context) error {
	m.mutex.Lock()
	if m.stopped {
		defer m.mutex.Unlock()
		return ErrMgrAlreadyStop
	}
	m.wg.Add(1)
	m.mutex.Unlock()

	go func() {
		defer tools.PrintPanicStack()
		defer m.wg.Done()
		defer m.cleaner.Stop()
		defer m.runningMod.Stop()
		defer m.deletingMod.Stop()

		m.log.Infof("async mgr run")
		// run components
		if err := m.cleaner.Run(ctx); err != nil {
			m.log.Errorf("async mgr run clear err: %v", err)
			return
		}
		if err := m.runningMod.Run(ctx); err != nil {
			m.log.Errorf("async mgr run running module err: %v", err)
			return
		}
		if err := m.deletingMod.Run(ctx); err != nil {
			m.log.Errorf("async mgr run deleting module err: %v", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stop:
				return
			case _, ok := <-m.runningMod.Need():
				if !ok {
					m.log.Errorf("async mgr running module stop")
					return
				}
				if dbTask, err := m.client.SelectOneRun(ctx, m.getTaskTypes()); err != nil {
					m.log.Errorf("async mgr select pending run task err: %v", err)
				} else if dbTask != nil {
					if task, err := m.newTask(dbTask); err != nil {
						m.log.Errorf("async mgr run task err: %v", err)
					} else {
						m.runningMod.RunTask(ctx, task)
					}
				}
			case _, ok := <-m.deletingMod.Need():
				if !ok {
					m.log.Errorf("async mgr deleting module stop")
					return
				}
				if dbTask, err := m.client.SelectOneDelete(ctx, m.getTaskTypes()); err != nil {
					m.log.Errorf("async mgr select pending del task err: %v", err)
				} else if dbTask != nil {
					if task, err := m.newTask(dbTask); err != nil {
						m.log.Errorf("async mgr delete task err: %v", err)
					} else {
						m.deletingMod.DeleteTask(ctx, task)
					}

				}
			}
		}
	}()
	return nil
}

func (m *Mgr) Stop() {
	m.mutex.Lock()
	// check stop
	if m.stopped {
		defer m.mutex.Unlock()
		m.log.Errorf(ErrMgrAlreadyStop.Error())
		return
	}
	// stop
	m.stopped = true
	m.stop <- struct{}{}
	m.mutex.Unlock()
	// wait
	m.wg.Wait()
	m.log.Infof("async mgr stop")
}

func (m *Mgr) RegisterTask(taskTyp uint32, newTask async_task.ITaskFunc) error {
	if m.checkStop() {
		return ErrMgrAlreadyStop
	}
	if newTask == nil {
		return fmt.Errorf("taskTyp %v newTask nil", taskTyp)
	}
	if _, ok := m.taskTypes.LoadOrStore(taskTyp, newTask); ok {
		return fmt.Errorf("taskTyp %v already registered", taskTyp)
	}
	return nil
}

func (m *Mgr) CreateTask(ctx context.Context, user, group string, taskTyp uint32, taskCtx string, autoRun bool) (uint32, error) {
	if m.checkStop() {
		return 0, ErrMgrAlreadyStop
	}
	if _, ok := m.taskTypes.Load(taskTyp); !ok {
		return 0, fmt.Errorf("taskTyp %v not registered", taskTyp)
	}
	return m.client.CreateTask(ctx, user, group, taskTyp, taskCtx, autoRun)
}

func (m *Mgr) ChangeTaskGroup(ctx context.Context, taskID uint32, group string) error {
	if m.checkStop() {
		return ErrMgrAlreadyStop
	}
	return m.client.ChangeTaskGroup(ctx, taskID, group)
}

func (m *Mgr) UserRun(ctx context.Context, taskID uint32) error {
	if m.checkStop() {
		return ErrMgrAlreadyStop
	}
	return m.client.TransStatus(ctx, taskID, trans.EventUserRun)
}

func (m *Mgr) UserDelete(ctx context.Context, taskID uint32) error {
	if m.checkStop() {
		return ErrMgrAlreadyStop
	}
	return m.client.TransStatus(ctx, taskID, trans.EventUserDelete)
}

func (m *Mgr) UserPause(ctx context.Context, taskID uint32) error {
	if m.checkStop() {
		return ErrMgrAlreadyStop
	}
	return m.client.TransStatus(ctx, taskID, trans.EventUserPause)
}

func (m *Mgr) GetTask(ctx context.Context, taskID uint32) (*model.AsyncTask, error) {
	if m.checkStop() {
		return nil, ErrMgrAlreadyStop
	}
	return m.client.GetTask(ctx, taskID)
}

func (m *Mgr) GetTasks(ctx context.Context, user, group string, taskTypes []uint32, status []trans.TaskStatus, offset, limit int32) ([]*model.AsyncTask, error) {
	if m.checkStop() {
		return nil, ErrMgrAlreadyStop
	}
	return m.client.GetTasks(ctx, user, group, taskTypes, status, offset, limit)
}

func (m *Mgr) checkStop() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.stopped
}

func (m *Mgr) getTaskTypes() []uint32 {
	var taskTypes []uint32
	m.taskTypes.Range(func(taskType, _ interface{}) bool {
		taskTypes = append(taskTypes, taskType.(uint32))
		return true
	})
	return taskTypes
}

func (m *Mgr) newTask(dbTask *model.AsyncTask) (*task.Task, error) {
	newTask, ok := m.taskTypes.Load(dbTask.Type)
	if !ok {
		return nil, fmt.Errorf("async task %v type %v not registered", dbTask.ID, dbTask.Type)
	}
	return task.NewTask(dbTask, newTask.(async_task.ITaskFunc)(), m.log), nil
}

var (
	ErrMgrAlreadyStop = errors.New("async mgr already stop")
)
