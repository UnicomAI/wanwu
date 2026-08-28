package store

import (
	"sync"

	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/model"
)

// todo 目前只是用数组存储，需要压测并考虑性能问题
type MemoryStore struct {
	mu       sync.RWMutex
	messages map[string][]*model.Message       // sessionID -> messages
	extMap   map[string]map[string]interface{} //扩展信息
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		messages: make(map[string][]*model.Message),
		extMap:   make(map[string]map[string]interface{}),
	}
}

func (s *MemoryStore) AddMessage(msg *model.Message, userSession *model.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionID := userSession.SessionID()

	s.messages[sessionID] = append(s.messages[sessionID], msg)
	return nil
}

// AddExtMessage 添加会话扩展信息
func (s *MemoryStore) AddExtMessage(extMap map[string]interface{}, userSession *model.Session) error {
	if len(extMap) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionID := userSession.SessionID()
	extMapTemp := s.extMap[sessionID]
	if extMapTemp == nil {
		extMapTemp = make(map[string]interface{})
		s.extMap[sessionID] = extMapTemp
	}
	for key, value := range extMap {
		extMapTemp[key] = value
	}
	return nil
}

func (s *MemoryStore) CompactMessage(msg *model.Message, userSession *model.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionID := userSession.SessionID()

	messages := s.messages[sessionID]
	if len(messages) > 0 {
		messages[len(messages)-1] = msg
	} else {
		messages = append(messages, msg)
	}
	s.messages[sessionID] = messages
	return nil
}

func (s *MemoryStore) GetMessages(userSession *model.Session) ([]*model.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessionID := userSession.SessionID()

	msgs := s.messages[sessionID]
	if msgs == nil {
		return []*model.Message{}, nil
	}

	result := make([]*model.Message, len(msgs))
	copy(result, msgs)
	return result, nil
}

// GetExtMessage 查询会话扩展信息
func (s *MemoryStore) GetExtMessage(userSession *model.Session) map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessionID := userSession.SessionID()

	return s.extMap[sessionID]
}
func (s *MemoryStore) GetCurrentMessage(userSession *model.Session) (*model.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessionID := userSession.SessionID()

	msgs := s.messages[sessionID]
	if len(msgs) > 0 {
		message := msgs[len(msgs)-1]
		return message.Copy(), nil
	}
	return nil, nil
}

func (s *MemoryStore) DeleteSession(userSession *model.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionID := userSession.SessionID()

	delete(s.messages, sessionID)
	delete(s.extMap, sessionID)
	return nil
}
