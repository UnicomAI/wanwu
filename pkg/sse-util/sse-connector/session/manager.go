package session

import (
	"context"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/model"
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/store"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	SessionMaxTime = 2 * time.Hour
)

// Subscriber 订阅者，代表一个 SSE 连接
type Subscriber struct {
	Chan chan *model.Message
}

// Manager 会话管理器
type Manager struct {
	Ctx     context.Context
	cancel  context.CancelFunc
	Invalid bool //会话失效，目前是clientID 或者 conversationID 为空,为了简化接入流程，所以参数不符合预期目前不报错，只设置invalid为true

	userSession *model.Session
	store       store.MessageStore
	callback    func(sessionId string)

	mu         sync.RWMutex
	subscriber *Subscriber
	writeDone  bool //是否已写完
}

func NewManager(ctx context.Context, s store.MessageStore, userSession *model.Session, callback func(sessionId string)) *Manager {
	detachContext := trace_util.DetachContext(ctx)
	ctx, cancel := context.WithCancel(detachContext)
	return &Manager{
		Ctx:         ctx,
		cancel:      cancel,
		store:       s,
		userSession: userSession,
		callback:    callback,
	}
}

func (m *Manager) GetSession() *model.Session {
	return m.userSession
}

func (m *Manager) GetBgContext() context.Context {
	return m.Ctx
}

// AddExt 添加扩展信息
func (m *Manager) AddExt(extMap map[string]interface{}) {
	if m.Invalid {
		return
	}
	_ = m.store.AddExtMessage(extMap, m.userSession)
}

// GetExt 查询扩展信息
func (m *Manager) GetExt() map[string]interface{} {
	if m.Invalid {
		return nil
	}
	return m.store.GetExtMessage(m.userSession)
}

// GetHistory 获取历史消息
func (m *Manager) GetHistory() ([]*model.Message, error) {
	if m.Invalid {
		return make([]*model.Message, 0), nil
	}
	return m.store.GetMessages(m.userSession)
}

// InvalidManager 将会话标记为无效，使 Publish/Subscribe 等方法成为 no-op。
// 仅可在 NewSSESession 返回前调用（构造阶段），调用后 Invalid 字段不再修改，因此无需加锁。
func (m *Manager) InvalidManager() {
	m.Invalid = true
}

// Subscribe 订阅会话消息
func (m *Manager) Subscribe() *Subscriber {
	if m.Invalid {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.writeDone {
		return nil
	}
	// 关闭已有 subscriber，避免已有连接停止输出但未关闭（新赶旧策略，保证断线重连不失败）
	if m.subscriber != nil && m.subscriber.Chan != nil {
		close(m.subscriber.Chan)
		m.subscriber = nil
	}
	sub := &Subscriber{Chan: make(chan *model.Message, 128)}
	m.subscriber = sub

	go m.DelayUnsubscribe(SessionMaxTime)
	return sub
}

// Unsubscribe 取消订阅
func (m *Manager) Unsubscribe() {
	if m.Invalid {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	subscriber := m.subscriber
	if subscriber == nil {
		return
	}
	if subscriber.Chan != nil {
		close(subscriber.Chan)
		m.subscriber = nil
	}
}

// DelayUnsubscribe 延迟清理会话
func (m *Manager) DelayUnsubscribe(delay time.Duration) {
	time.Sleep(delay)
	m.Unsubscribe()
}

// Publish 发布消息给所有订阅者
func (m *Manager) Publish(msg *model.Message, compactProcessor func(currentMsg *model.Message, lastMsg *model.Message) (bool, *model.Message)) error {
	if m.Invalid {
		return nil
	}
	//生成消息ID,保证id 递增
	msg.ID = util.NewID()

	if compactProcessor != nil { //处理消息合并
		lastMsg, err := m.store.GetCurrentMessage(m.userSession)
		if err != nil {
			return err
		}
		var skip = true
		var compactMsg *model.Message
		if lastMsg != nil {
			skip, compactMsg = compactProcessor(msg, lastMsg)
		}
		if skip { // 存储消息
			err = m.store.AddMessage(msg, m.userSession)
			if err != nil {
				return err
			}
		} else {
			err = m.store.CompactMessage(compactMsg, m.userSession)
			if err != nil {
				return err
			}
		}
	} else { // 存储消息
		err := m.store.AddMessage(msg, m.userSession)
		if err != nil {
			return err
		}
	}

	//发送消息
	m.mu.RLock()
	subscriber := m.subscriber
	m.mu.RUnlock()
	if subscriber != nil {
		//发送消息
		select {
		case subscriber.Chan <- msg:
		default:
			log.Warnf("Publish %s subscriber channel full, dropping message", m.userSession.SessionID())
		}
	}
	return nil
}

// Cancel 终止会话：取消后端执行、关闭订阅者 channel、清理会话状态。
func (m *Manager) Cancel() error {
	if m.Invalid {
		return nil
	}
	if m.cancel != nil {
		m.cancel()
	}
	m.Unsubscribe()
	return m.finish()
}

// finish 清理会话状态：标记写入完成、删除存储、从注册表移除。writeDone防重入。
func (m *Manager) finish() error {
	if m.Invalid {
		return nil
	}

	m.mu.Lock()
	if m.writeDone {
		m.mu.Unlock()
		return nil
	}
	m.writeDone = true
	m.mu.Unlock()

	if m.callback != nil {
		m.callback(m.userSession.SessionID())
	}
	return m.store.DeleteSession(m.userSession)
}
