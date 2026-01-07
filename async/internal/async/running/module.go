package running

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
	RunTask(ctx context.Context, task *task.Task)
}

type runMod struct {
	log    async_config.Logger
	client client.ITaskClient

	maxConcurrency int
	concurrency    chan struct{}

	needInterval int // second
	need         chan struct{}

	checkInterval int // second

	heartbeatInterval int // second

	mutex   sync.Mutex
	tasks   sync.Map // taskID -> task
	stopped bool
	stop    chan struct{}

	wg sync.WaitGroup
}

func NewModule(c client.ITaskClient, cfg config.Config) IModule {
	return &runMod{
		log:               cfg.Log,
		client:            c,
		maxConcurrency:    cfg.RunMaxConcurrency,
		concurrency:       make(chan struct{}, cfg.RunMaxConcurrency),
		needInterval:      cfg.RunTaskInterval,
		need:              make(chan struct{}, 1),
		checkInterval:     config.RunCheckInterval,
		heartbeatInterval: config.TaskHeartbeatInterval,
		stop:              make(chan struct{}, 1),
	}
}

func (m *runMod) Run(ctx context.Context) error {
	m.mutex.Lock()
	if m.stopped {
		defer m.mutex.Unlock()
		return errors.New("async running module already stop")
	}
	m.wg.Add(1)
	m.mutex.Unlock()

	go func() {
		defer tools.PrintPanicStack()
		defer m.wg.Done()

		m.log.Infof("async running module run")
		needTicker := time.NewTicker(time.Duration(m.needInterval) * time.Second)
		defer needTicker.Stop()
		checkTicker := time.NewTicker(time.Duration(m.checkInterval) * time.Second)
		defer checkTicker.Stop()
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
			case <-checkTicker.C:
				m.tasks.Range(func(taskID, t interface{}) bool {
					if ok, err := m.client.CheckStop(ctx, taskID.(uint32)); err != nil {
						m.log.Errorf("async running module check task %v stop err: %v", taskID.(uint32), err)
					} else if ok {
						if err := t.(*task.Task).SendStop(); err != nil {
							m.log.Errorf("async running module send stop err: %v", err)
						}
					}
					return true
				})
			}
		}
	}()
	return nil
}

func (m *runMod) Stop() {
	m.mutex.Lock()
	// check stop
	if m.stopped {
		defer m.mutex.Unlock()
		m.log.Errorf("async running module already stop")
		return
	}
	// stop
	m.stopped = true
	m.stop <- struct{}{}
	m.tasks.Range(func(_, t interface{}) bool {
		if err := t.(*task.Task).SendStop(); err != nil {
			m.log.Errorf("async running module send stop err: %v", err)
		}
		return true
	})
	m.mutex.Unlock()
	// wait
	m.wg.Wait()
	m.log.Infof("async running module stop")
}

func (m *runMod) Need() <-chan struct{} {
	return m.need
}

func (m *runMod) RunTask(ctx context.Context, task *task.Task) {
	m.mutex.Lock()
	// check stop
	if m.stopped {
		defer m.mutex.Unlock()
		m.log.Errorf("async running module run task %v err: async running module already stop", task.ID())
		return
	}
	// check concurrency
	select {
	case m.concurrency <- struct{}{}:
	default:
		defer m.mutex.Unlock()
		m.log.Errorf("async running module run task %v err: max concurrency", task.ID())
		return
	}
	// add task
	if _, ok := m.tasks.LoadOrStore(task.ID(), task); ok {
		defer m.mutex.Unlock()
		m.log.Errorf("async running module run task %v err: already exist", task.ID())
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
		reportCh, err := task.Running(ctx)
		if err != nil {
			m.log.Errorf("async running module run task %v err: %v", task.ID(), err)
			return
		}
		m.log.Debugf("async task %v start running, initCtx: %v", task.ID(), task.InitCtx())

		ticker := time.NewTicker(time.Duration(m.heartbeatInterval) * time.Second)
		defer ticker.Stop()
		var stopped bool
		for {
			select {
			case <-ticker.C:
				if err := m.client.UpdateHeartbeat(ctx, task.ID()); err != nil {
					m.log.Errorf("async task %v running heartbeat err: %v", task.ID(), err)
				} else {
					m.log.Debugf("async task %v running heartbeat", task.ID())
				}
			case report, ok := <-reportCh:
				// check stop
				if !ok {
					stopped = true
					break
				}
				// check report
				if report == nil {
					m.log.Errorf("async task %v running report nil", task.ID())
					continue
				}
				currentPhase, needDelete := report.Phase()
				eventCtx := report.Context()
				m.log.Debugf("async task %v running report phase %v event ctx %v", task.ID(), currentPhase, eventCtx)
				// update context
				if err := m.client.UpdateContext(ctx, task.ID(), eventCtx); err != nil {
					m.log.Errorf("async task %v running report phase %v event ctx %v update err: %v", task.ID(), currentPhase, eventCtx, err)
				}
				// update state
				if phase != async_task.RunPhaseNormal {
					m.log.Errorf("async task %v running report phase %v err: current phase %v not normal", task.ID(), currentPhase, phase)
				}
				if currentPhase == async_task.RunPhaseFinished || currentPhase == async_task.RunPhaseFailed {
					phase = currentPhase
					// delete
					if needDelete {
						if err := m.client.Delete(ctx, task.ID()); err != nil {
							m.log.Errorf("async task %v running report phase %v delete err: %v", task.ID(), phase, err)
						} else {
							m.log.Debugf("async task %v running report phase %v delete", task.ID(), phase)
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
						m.log.Errorf("async task %v running report phase %v trans event %v err: %v", task.ID(), phase, event, err)
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
					m.log.Errorf("async task %v puase running phase normal err: %v", task.ID(), err)
				} else {
					m.log.Debugf("async task %v pause running phase normal", task.ID())
				}
			} else {
				// task panic, auto failed later
				m.log.Errorf("async task %v stop running phase normal maybe panic", task.ID())
			}
		case async_task.RunPhaseFinished:
			m.log.Debugf("async task %v stop running phase finished", task.ID())
		case async_task.RunPhaseFailed:
			m.log.Errorf("async task %v stop running phase failed", task.ID())
		default:
		}
	}()
}
