package types

import (
	"context"
	"errors"
)

// StreamSender 流式回复发送器，由平台适配器创建，用于逐 chunk 发送流式内容
// 平台适配器（如钉钉）可以实现此接口以支持流式卡片更新
type StreamSender interface {
	// SendChunk 发送一个流式内容块
	// content: 增量内容（本次新增的文本）
	// isFinal: 是否为最后一个块
	SendChunk(ctx context.Context, content string, isFinal bool) error
	// SetContent 用完整内容替换卡片正文（非追加）。
	// 用于把聚合后的完整 markdown 过程体推给卡片：每个事件后整体重渲染。
	// isFinal=true 时将卡片置 finished。平台不支持流式时 CreateStreamSender 返回 nil，
	// 调用方应降级为 SendMessage，不会调用本方法。
	SetContent(ctx context.Context, content string, isFinal bool) error
	// Close 收尾流式发送器，必须在所有退出路径上调用（成功/失败/取消）。
	// err==nil 表示正常完成（卡片置 finished）；err!=nil 表示失败（卡片置 failed）。
	// 必须幂等：重复调用不应有副作用。
	Close(ctx context.Context, err error) error
}

// Adapter 平台适配器接口
// 每个平台（钉钉/微信/飞书）实现此接口
type Adapter interface {
	// Connect 连接到平台
	Connect(config AdapterConfig) error
	// Disconnect 断开平台连接
	Disconnect() error
	// IsConnected 检查是否已连接
	IsConnected() bool
	// GetAccountInfo 获取平台账号信息
	GetAccountInfo() (accountId, nickname, avatar string, err error)
	// SendMessage 向平台用户发送消息
	SendMessage(ctx context.Context, userID string, content string, extra map[string]string) error
	// OnMessage 注册消息回调
	OnMessage(handler MessageHandler)
	// CreateStreamSender 创建流式回复发送器
	// 返回 nil 表示该平台不支持流式回复，调用方应降级为 SendMessage
	CreateStreamSender(ctx context.Context, userID string, extra map[string]string) StreamSender
}

// MessageHandler 平台消息回调处理函数
type MessageHandler func(ctx context.Context, msg *PlatformMessage) error

// PlatformMessage 平台消息统一格式
type PlatformMessage struct {
	ChannelID      string            // 通道 ID
	ConversationID string            // 平台会话 ID（用于维持上下文）
	UserID         string            // 平台用户 ID
	Content        string            // 消息内容
	MsgType        string            // 消息类型：text/image/markdown/...
	ChannelType    string            // 平台类型：dingtalk/wechat/feishu
	Extra          map[string]string // 额外信息（如 sessionWebhook 等）
	Attachments    []Attachment      // 附件（图片/文件等，已下载为字节）
}

// Attachment 平台附件（图片/文件等）
// 适配器在 inbound 收到图片/文件消息时下载为字节填入 Data，
// chat handler 上传到万悟 minio 后作为 WGA 多模态 binary.url 发给通用智能体。
type Attachment struct {
	Name     string // 文件名
	MimeType string // MIME 类型
	Data     []byte // 文件字节
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	ChannelID   string            // 通道 ID
	ChannelType string            // 平台类型：dingtalk/wechat/feishu
	AppKey      string            // 钉钉 appKey
	AppSecret   string            // 钉钉 appSecret
	Token       string            // 微信 token
	BaseUrl     string            // 微信 baseUrl
	AccountId   string            // 微信 accountId
	AppId       string            // 飞书 appId
	EncryptKey  string            // 飞书 encryptKey
	VerifyToken string            // 飞书 verificationToken
	ConnMode    string            // 连接模式：stream/webhook（钉钉）或 websocket/webhook（飞书）
	Extra       map[string]string // 额外配置
}

// ChannelType 常量
const (
	ChannelTypeDingTalk = "dingtalk"
	ChannelTypeWeChat   = "wechat"
	ChannelTypeFeiShu   = "feishu"
)

// WebhookHandler Webhook 回调处理接口
// 支持 Webhook 模式的平台适配器可以实现此接口
type WebhookHandler interface {
	// HandleWebhook 处理平台 Webhook 回调请求
	HandleWebhook(ctx context.Context, body []byte, timestamp, sign string) error
}

// FileSender 文件发送能力接口（可选实现）。
// 平台适配器可选实现此接口以支持发送文件附件（如 WGA 工作区产物回发 IM）。
// 钉钉/微信均已实现（钉钉支持普通上传 ≤20MB + 分块事务上传 >20MB）；飞书暂不实现，
// Manager.SendFile 会返回 ErrFileSendUnsupported，由上层降级为文本提示。
type FileSender interface {
	SendFile(ctx context.Context, userID, fileName, mimeType string, data []byte, extra map[string]string) error
}

// ErrFileSendUnsupported 当前平台适配器不支持发送文件。
// Manager.SendFile 在适配器未实现 FileSender 时返回，调用方据此降级为文本提示。
var ErrFileSendUnsupported = errors.New("file send not supported on this platform")
