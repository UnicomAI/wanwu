package client

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
)

// IClient 通道服务数据访问接口
type IClient interface {
	// --- Channel ---
	CreateChannel(ctx context.Context, channel *model.Channel) (*model.Channel, error)
	GetChannel(ctx context.Context, channelID string) (*model.Channel, error)
	ListChannels(ctx context.Context, userID, orgID, name string, pageNo, pageSize int32) ([]*model.Channel, int64, error)
	UpdateChannel(ctx context.Context, channelID string, updates map[string]interface{}) (*model.Channel, error)
	DeleteChannel(ctx context.Context, channelID string) error
	ListEnabledChannels(ctx context.Context) ([]*model.Channel, error)

	// --- QR Session ---
	CreateQRSession(ctx context.Context, session *model.QRSession) (*model.QRSession, error)
	GetQRSession(ctx context.Context, sessionID string) (*model.QRSession, error)
	UpdateQRSession(ctx context.Context, sessionID string, updates map[string]interface{}) error
	DeleteQRSession(ctx context.Context, sessionID string) error
	DeleteExpiredQRSessions(ctx context.Context, now int64) (int64, error)

	// --- ChannelConversation 会话映射（持久化 threadId/conversationId） ---
	GetConversation(ctx context.Context, channelID, userID, appType string) (*model.ChannelConversation, error)
	ListConversationsByChannel(ctx context.Context, channelID string, limit int) ([]*model.ChannelConversation, error)
	UpsertConversation(ctx context.Context, conv *model.ChannelConversation) error
	DeleteConversationsByChannel(ctx context.Context, channelID string) error
}
