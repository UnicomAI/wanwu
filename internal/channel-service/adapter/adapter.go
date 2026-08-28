package adapter

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
)

// Re-export types from types package for backward compatibility
type Adapter = types.Adapter
type MessageHandler = types.MessageHandler
type PlatformMessage = types.PlatformMessage
type AdapterConfig = types.AdapterConfig
type WebhookHandler = types.WebhookHandler

const (
	ChannelTypeDingTalk = types.ChannelTypeDingTalk
	ChannelTypeWeChat   = types.ChannelTypeWeChat
	ChannelTypeFeiShu   = types.ChannelTypeFeiShu
)

// Keep Adapter interface documentation here
//
// Adapter 平台适配器接口
// 每个平台（钉钉/微信/飞书）实现此接口
// Connect(config AdapterConfig) error
// Disconnect() error
// IsConnected() bool
// GetAccountInfo() (accountId, nickname, avatar string, err error)
// SendMessage(ctx context.Context, userID string, content string) error
// OnMessage(handler MessageHandler)

// WebhookHandler Webhook 回调处理接口
// 支持 Webhook 模式的平台适配器可以实现此接口
// HandleWebhook(ctx context.Context, body []byte, timestamp, sign string) error

// Suppress unused import warning
var _ context.Context
