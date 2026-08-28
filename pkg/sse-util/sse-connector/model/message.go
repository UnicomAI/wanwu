package model

import "fmt"

type Message struct {
	ID   string      `json:"ID"`
	Data interface{} `json:"data"`
}

func (m *Message) Copy() *Message {
	return &Message{
		ID:   m.ID,
		Data: m.Data,
	}
}

type Session struct {
	ConversationID string `json:"conversationId"`
	ClientID       string `json:"clientId"`
}

func (s *Session) SessionID() string {
	return fmt.Sprintf("%s-%s", s.ConversationID, s.ClientID)
}

func (s *Session) Check() bool {
	return len(s.ConversationID) > 0 && len(s.ClientID) > 0
}
