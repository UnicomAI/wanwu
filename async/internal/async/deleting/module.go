package deleting

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/async/internal/async/config"
	"github.com/UnicomAI/wanwu/async/internal/async/task"
	"github.com/UnicomAI/wanwu/async/internal/db/client"
	"github.com/UnicomAI/wanwu/async/internal/db/trans"
	"github.com/UnicomAI/wanwu/async/internal/tools"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_task"
)

// IModule goroutine safe
type IModule interface {
	Run(ctx context.Context) error
	Stop()

	Need() <-chan struct{}
	DeleteTask(ctx context.Context, task *task.Task)
}

type delMod struct {
	log    async_config.Logger
	client client.ITaskClient

	maxConcurrency int
	concurrency    chan struct{}

	needInterval int // second
	need         chan struct{}

	heartbeatInterval int // second

	mutex   sync.Mutex
	tasks   sync.Map // taskID -> task
	stopped bool
	stop    chan struct{}

	wg sync.WaitGroup
}

func NewModule(c client.ITaskClient, cfg config.Config) IModule {
	return &delMod{
		log:               cfg.Log,
		client:            c,
		maxConcurrency:    config.DeleteMaxConcurrency,
		concurrency:       make(chan struct{}, config.DeleteMaxConcurrency),
		needInterval:      config.DeleteTaskInterval,
		need:              make(chan struct{}, 1),
		heartbeatInterval: config.TaskHeartbeatInterval,
		stop:              make(chan struct{}, 1),
	}
}

func (m *delMod) Run(ctx context.Context) error {
	m.mutex.Lock()
	if m.stopped {
		defer m.mutex.Unlock()
		return errors.New("async deleting module already stop")
	}
	m.wg.Add(1)
	m.mutex.Unlock()

	go func() {
		defer tools.PrintPanicStack()
		defer m.wg.Done()

		m.log.Infof("async deleting module run")
		needTicker := time.NewTicker(time.Duration(m.needInterval) * time.Second)
		defer needTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stop:
				return
			case <-needTicker.C:
				if len(m.concurrency) >= m.maxConcurrency {
					continue
				}
				select {
				case m.need <- struct{}{}:
				default: // do nothing
				}
			}
		}
	}()
	return nil
}

func (m *delMod) Stop() {
	m.mutex.Lock()
	// check stop
	if m.stopped {
		defer m.mutex.Unlock()
		m.log.Errorf("async deleting module already stop")
		return
	}
	// stop
	m.stopped = true
	m.stop <- struct{}{}
	m.tasks.Range(func(_, t interface{}) bool {
		if err := t.(*task.Task).SendStop(); err != nil {
			m.log.Errorf("async deleting module send stop err: %v", err)
		}
		return true
	})
	m.mutex.Unlock()
	// wait
	m.wg.Wait()
	m.log.Infof("async deleting module stop")
}

func (m *delMod) Need() <-chan struct{} {
	return m.need
}

func (m *delMod) DeleteTask(ctx context.Context, task *task.Task) {
	m.mutex.Lock()
	// check stop
	if m.stopped {
		defer m.mutex.Unlock()
		m.log.Errorf("async deleting module delete task %v err: async deleting module already stop", task.ID())
		return
	}
	// check concurrency
	select {
	case m.concurrency <- struct{}{}:
	default:
		defer m.mutex.Unlock()
		m.log.Errorf("async deleting module delete task %v err: max concurrency", task.ID())
		return
	}
	// add task
	if _, ok := m.tasks.LoadOrStore(task.ID(), task); ok {
		defer m.mutex.Unlock()
		m.log.Errorf("async deleting module delete task %v err: already exist", task.ID())
		return
	}
	m.wg.Add(1)
	m.mutex.Unlock()

	go func() {
		defer tools.PrintPanicStack()
		defer m.wg.Done()
		defer func() {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			m.tasks.Delete(task.ID())
		}()
		defer func() {
			<-m.concurrency
		}()

		phase := async_task.RunPhaseNormal
		reportCh, err := task.Deleting(ctx)
		if err != nil {
			m.log.Errorf("async deleting module delete task %v err: %v", task.ID(), err)
			return
		}
		m.log.Debugf("async task %v start deleting", task.ID())

		ticker := time.NewTicker(time.Duration(m.heartbeatInterval) * time.Second)
		defer ticker.Stop()
		var stopped bool
		for {
			select {
			case <-ticker.C:
				if err := m.client.UpdateHeartbeat(ctx, task.ID()); err != nil {
					m.log.Errorf("async task %v deleting heartbeat err: %v", task.ID(), err)
				} else {
					m.log.Debugf("async task %v deleting heartbeat", task.ID())
				}
			case report, ok := <-reportCh:
				// check stop
				if !ok {
					stopped = true
					break
				}
				// check report
				if report == nil {
					m.log.Errorf("async task %v deleting report nil", task.ID())
					continue
				}
				currentPhase, needDelete := report.Phase()
				eventCtx := report.Context()
				m.log.Debugf("async task %v deleting report phase %v event ctx %v", task.ID(), currentPhase, eventCtx)
				// update context
				if err := m.client.UpdateContext(ctx, task.ID(), eventCtx); err != nil {
					m.log.Errorf("async task %v deleting report phase %v event ctx %v err: %v", task.ID(), currentPhase, eventCtx, err)
				}
				// update state
				if currentPhase == async_task.RunPhaseFinished || currentPhase == async_task.RunPhaseFailed {
					phase = currentPhase
					// delete
					if needDelete {
						if err := m.client.Delete(ctx, task.ID()); err != nil {
							m.log.Errorf("async task %v deleting report phase %v delete err: %v", task.ID(), phase, eventCtx)
						} else {
							m.log.Debugf("async task %v deleting report phase %v delete", task.ID(), phase)
						}
						return
					}
					// update state
					var event trans.TaskEvent
					switch phase {
					case async_task.RunPhaseFinished:
						event = trans.EventTaskFinished
					case async_task.RunPhaseFailed:
						event = trans.EventTaskFailed
					}
					if err := m.client.TransStatus(ctx, task.ID(), event); err != nil {
						m.log.Errorf("async task %v deleting report phase %v trans event %v err: %v", task.ID(), phase, event, err)
					}
				}
			}
			if stopped {
				break
			}
		}

		switch phase {
		case async_task.RunPhaseNormal:
			if task.CheckStop() {
				// sys pause
				if err := m.client.TransStatus(ctx, task.ID(), trans.EventSysPause); err != nil {
					m.log.Errorf("async task %v puase deleting phase normal err: %v", task.ID(), err)
				} else {
					m.log.Debugf("async task %v pause deleting phase normal", task.ID())
				}
			} else {
				// task panic, auto failed later
				m.log.Errorf("async task %v stop deleting phase normal maybe panic", task.ID())
			}
		case async_task.RunPhaseFinished:
			m.log.Debugf("async task %v stop deleting phase finished", task.ID())
		case async_task.RunPhaseFailed:
			m.log.Errorf("async task %v stop deleting phase failed", task.ID())
		default:
		}
	}()
}
