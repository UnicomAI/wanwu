package store

import (
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/model"
)

// MessageStore 存储接口，支持后续扩展持久化
type MessageStore interface {
	// AddMessage 添加消息到会话
	AddMessage(msg *model.Message, userSession *model.Session) error

	// AddExtMessage 添加会话扩展信息
	AddExtMessage(extMap map[string]interface{}, userSession *model.Session) error

	// CompactMessage 合并消息到会话,如果当前会话有数据则用此数据覆盖最新一条，如果没有数据则添加一条
	CompactMessage(msg *model.Message, userSession *model.Session) error

	// GetMessages 获取会话的所有消息
	GetMessages(userSession *model.Session) ([]*model.Message, error)

	// GetExtMessage 查询会话扩展信息
	GetExtMessage(userSession *model.Session) map[string]interface{}

	// GetCurrentMessage 获取会话的当前最新消息
	GetCurrentMessage(userSession *model.Session) (*model.Message, error)

	// DeleteSession 删除会话及其消息
	DeleteSession(userSession *model.Session) error
}
