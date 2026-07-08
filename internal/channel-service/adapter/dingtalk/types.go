package dingtalk

import (
	"context"
	"time"
)

// MessageHandler 消息处理器
type MessageHandler func(ctx context.Context, msg *Message) error

// MessageType 消息类型
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeMarkdown MessageType = "markdown"
	MessageTypeImage    MessageType = "picture"
	MessageTypeVoice    MessageType = "voice"
	MessageTypeFile     MessageType = "file"
	MessageTypeLink     MessageType = "link"
	MessageTypeAction   MessageType = "actionCard"
)

// Message 钉钉统一消息格式
type Message struct {
	MsgType        MessageType            `json:"msgtype"`
	Content        string                 `json:"content,omitempty"`
	Conversation   string                 `json:"conversation,omitempty"` // 群会话 ID
	Sender         string                 `json:"sender,omitempty"`       // 发送者 ID（staffId 优先）
	SenderNick     string                 `json:"sender_nick,omitempty"`  // 发送者昵称
	Receiver       string                 `json:"receiver,omitempty"`     // 接收者 ID（机器人 ID）
	ChannelID      string                 `json:"channel_id,omitempty"`   // 渠道 ID（用于路由）
	MessageID      string                 `json:"message_id,omitempty"`   // 钉钉消息 ID（用于去重）
	Timestamp      time.Time              `json:"timestamp"`
	IsInGroup      bool                   `json:"is_in_group,omitempty"`
	AtUserIDs      []string               `json:"at_user_ids,omitempty"`
	SessionWebhook string                 `json:"session_webhook,omitempty"` // sessionWebhook 用于回复
	Raw            map[string]interface{} `json:"raw,omitempty"`             // 原始消息数据
	Attachments    []Attachment           `json:"attachments,omitempty"`     // 图片/文件附件（已下载为字节）
}

// Attachment 钉钉附件（图片/文件），已下载为字节
type Attachment struct {
	Name     string
	MimeType string
	Data     []byte
}

// TextMessage 文本消息
type TextMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

// MarkdownMessage Markdown 消息
type MarkdownMessage struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
}

// LinkMessage 链接消息
type LinkMessage struct {
	MsgType string `json:"msgtype"`
	Link    struct {
		Title      string `json:"title"`
		Text       string `json:"text"`
		MessageURL string `json:"messageUrl"`
		PicURL     string `json:"picUrl,omitempty"`
	} `json:"link"`
}

// ActionCardMessage 卡片消息
type ActionCardMessage struct {
	MsgType    string `json:"msgtype"`
	ActionCard struct {
		Title          string `json:"title"`
		Text           string `json:"text"`
		SingleTitle    string `json:"singleTitle,omitempty"`
		SingleURL      string `json:"singleURL,omitempty"`
		BtnOrientation string `json:"btnOrientation,omitempty"`
		BtnJSONList    []struct {
			Title     string `json:"title"`
			ActionURL string `json:"actionURL"`
		} `json:"btnJsonList,omitempty"`
	} `json:"actionCard"`
}

// WebhookRequest 钉钉 Webhook 推送消息格式
type WebhookRequest struct {
	Msgtype    string `json:"msgtype"`
	Msgid      string `json:"msgid"`
	CreateTime int64  `json:"createTime"`

	// 文本消息
	Text struct {
		Content string `json:"content"`
	} `json:"text,omitempty"`

	// 发送者信息
	SenderNick    string `json:"senderNick"`
	SenderCorpRid string `json:"senderCorpRid"`
	Sender        string `json:"sender"`

	// 会话信息
	ConversationID    string `json:"conversationId"`
	ConversationType  string `json:"conversationType"`
	ConversationTitle string `json:"conversationTitle"`
	IsInAtList        bool   `json:"isInAtList"`

	// @ 信息
	AtUser struct {
		DingtalkID string `json:"dingtalkId"`
	} `json:"atUser,omitempty"`

	// 群信息
	ChatbotUserID string `json:"chatbotUserId"`

	// 图片消息
	Picture struct {
		PicURL string `json:"picURL"`
	} `json:"picture,omitempty"`

	// 语音消息
	Voice struct {
		Duration int    `json:"duration"`
		Content  string `json:"content"`
	} `json:"voice,omitempty"`

	// 文件消息
	File struct {
		FileName string `json:"fileName"`
		FileURL  string `json:"fileUrl"`
		FileSize int    `json:"fileSize"`
	} `json:"file,omitempty"`

	// 富文本消息
	RichText struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	} `json:"richText,omitempty"`

	// @ 用户列表
	AtUsers []struct {
		DingtalkID string `json:"dingtalkId"`
	} `json:"atUsers,omitempty"`

	// sessionWebhook
	SessionWebhook            string `json:"sessionWebhook"`
	SessionWebhookExpiredTime int64  `json:"sessionWebhookExpiredTime"`
}

// AccessTokenResponse 获取 access token 的响应（新版 API Token，POST /v1.0/oauth2/accessToken）
type AccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

// OapiTokenResponse 获取旧版 OAPI access token 的响应（GET /gettoken）。
// 与 AccessTokenResponse 区别：字段名为下划线风格 access_token（非 accessToken），
// 用于媒体/文件上传接口的 access_token query 参数。
type OapiTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// SendMessageResponse 发送消息的响应
type SendMessageResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgID   string `json:"msgid,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	UserID string `json:"userid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Mobile string `json:"mobile"`
	Email  string `json:"email"`
}

// RobotInfo 机器人信息
type RobotInfo struct {
	RobotID   string `json:"robotId"`
	RobotName string `json:"robotName"`
	Avatar    string `json:"avatar"`
}

// --- 卡片相关类型 ---

// AICardStatus AI 卡片状态
type AICardStatus string

const (
	AICardStatusProcessing AICardStatus = "1" // 处理中
	AICardStatusInputing   AICardStatus = "2" // 输入中
	AICardStatusFinished   AICardStatus = "3" // 执行完成
	AICardStatusExecuting  AICardStatus = "4" // 执行中
	AICardStatusFailed     AICardStatus = "5" // 执行失败
)

// DefaultCardTemplateID 钉钉官方 AI Markdown 卡片模板 ID
const DefaultCardTemplateID = "382e4302-551d-4880-bf29-a30acfab2e71.schema"

// CreateCardRequest 创建卡片实例请求
type CreateCardRequest struct {
	CardTemplateID        string          `json:"cardTemplateId"`
	OutTrackID            string          `json:"outTrackId"`
	CardData              CardData        `json:"cardData"`
	CallbackType          string          `json:"callbackType"`
	ImGroupOpenSpaceModel *CardSpaceModel `json:"imGroupOpenSpaceModel,omitempty"`
	ImRobotOpenSpaceModel *CardSpaceModel `json:"imRobotOpenSpaceModel,omitempty"`
}

// CardData 卡片数据
type CardData struct {
	CardParamMap map[string]string `json:"cardParamMap"`
}

// CardSpaceModel 卡片空间模型
type CardSpaceModel struct {
	SupportForward bool              `json:"supportForward"`
	RobotCode      string            `json:"robotCode,omitempty"`
	AtUserIDs      map[string]string `json:"atUserIds,omitempty"`
	Recipients     []string          `json:"recipients,omitempty"`
}

// DeliverCardRequest 投放卡片请求
type DeliverCardRequest struct {
	OutTrackID              string        `json:"outTrackId"`
	UserIDType              int           `json:"userIdType"`
	OpenSpaceID             string        `json:"openSpaceId,omitempty"`
	ImGroupOpenDeliverModel *DeliverModel `json:"imGroupOpenDeliverModel,omitempty"`
	ImRobotOpenDeliverModel *DeliverModel `json:"imRobotOpenDeliverModel,omitempty"`
}

// DeliverModel 投放模型
type DeliverModel struct {
	RobotCode  string            `json:"robotCode,omitempty"`
	AtUserIDs  map[string]string `json:"atUserIds,omitempty"`
	Recipients []string          `json:"recipients,omitempty"`
	SpaceType  string            `json:"spaceType,omitempty"`
}

// CreateAndDeliverCardRequest 创建并投放卡片请求（一步到位）
type CreateAndDeliverCardRequest struct {
	CardTemplateID        string          `json:"cardTemplateId"`
	OutTrackID            string          `json:"outTrackId"`
	CardData              CardData        `json:"cardData"`
	CallbackType          string          `json:"callbackType"`
	OpenSpaceID           string          `json:"openSpaceId,omitempty"`
	ImGroupOpenSpaceModel *CardSpaceModel `json:"imGroupOpenSpaceModel,omitempty"`
	ImRobotOpenSpaceModel *CardSpaceModel `json:"imRobotOpenSpaceModel,omitempty"`

	// 投放相关字段
	UserIDType              int           `json:"userIdType,omitempty"`
	ImGroupOpenDeliverModel *DeliverModel `json:"imGroupOpenDeliverModel,omitempty"`
	ImRobotOpenDeliverModel *DeliverModel `json:"imRobotOpenDeliverModel,omitempty"`
}

// StreamingCardRequest 流式更新卡片请求
type StreamingCardRequest struct {
	OutTrackID string `json:"outTrackId"`
	GUID       string `json:"guid"`
	Key        string `json:"key"`
	Content    string `json:"content"`
	IsFull     bool   `json:"isFull"`
	IsFinalize bool   `json:"isFinalize"`
	IsError    bool   `json:"isError"`
}

// UpdateCardRequest 更新卡片内容请求
type UpdateCardRequest struct {
	OutTrackID string   `json:"outTrackId"`
	CardData   CardData `json:"cardData"`
}
