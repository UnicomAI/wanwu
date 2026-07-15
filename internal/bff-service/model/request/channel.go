package request

// --- 万悟平台代理 ---

// ListWanwuAgentsRequest 获取万悟智能体列表请求（复用探索广场逻辑）
type ListWanwuAgentsRequest struct {
	Name       string `form:"name" json:"name"`             // 搜索名称
	AppType    string `form:"appType" json:"appType"`       // agent / rag
	SearchType string `form:"searchType" json:"searchType"` // all / favorite / private / history
}

func (r *ListWanwuAgentsRequest) Check() error { return nil }

// --- 扫码登录 ---

// CreateQRLoginRequest 发起扫码登录请求
type CreateQRLoginRequest struct {
	ChannelType string `uri:"channelType" json:"channelType" binding:"required"` // wechat/dingtalk/feishu
}

func (r *CreateQRLoginRequest) Check() error { return nil }

// GetQRLoginStatusRequest 查询扫码状态请求
type GetQRLoginStatusRequest struct {
	ChannelType string `uri:"channelType" json:"channelType" binding:"required"`
	SessionID   string `uri:"sessionId" json:"sessionId" binding:"required"`
}

func (r *GetQRLoginStatusRequest) Check() error { return nil }

// CancelQRLoginRequest 取消扫码登录请求
type CancelQRLoginRequest struct {
	ChannelType string `uri:"channelType" json:"channelType" binding:"required"`
	SessionID   string `uri:"sessionId" json:"sessionId" binding:"required"`
}

func (r *CancelQRLoginRequest) Check() error { return nil }

// CompleteQRLoginRequest 完成扫码登录请求
type CompleteQRLoginRequest struct {
	ChannelType string `uri:"channelType" json:"channelType" binding:"required"`
	SessionID   string `uri:"sessionId" json:"sessionId" binding:"required"`
}

func (r *CompleteQRLoginRequest) Check() error { return nil }

// --- 通道管理 ---

// CreateChannelRequest 创建通道请求
type CreateChannelRequest struct {
	Name        string            `json:"name" binding:"required"`
	ChannelType string            `json:"channelType" binding:"required"` // wechat/dingtalk/feishu
	AppType     string            `json:"appType"`                        // agent（默认）/ wga（通用智能体）
	AppID       string            `json:"appId"`
	ApiKeyId    string            `json:"apiKeyId"`
	ApiKey      string            `json:"apiKey"`    // API Key 完整值，创建时传入
	ModelUuid   string            `json:"modelUuid"` // WGA 通道使用的模型 UUID
	AgentId     string            `json:"agentId"`   // WGA 通道绑定的子智能体 ID（直连该子智能体，跳过 Supervisor）
	Config      map[string]string `json:"config" binding:"required"`
}

func (r *CreateChannelRequest) Check() error { return nil }

// UpdateChannelRequest 更新通道请求
type UpdateChannelRequest struct {
	Name      string `json:"name"`
	AppID     string `json:"appId"`
	ApiKeyId  string `json:"apiKeyId"`
	ApiKey    string `json:"apiKey"`
	ModelUuid string `json:"modelUuid"` // WGA 通道使用的模型 UUID
	// AgentId 用指针以区分三态（JSON 字段缺失→nil，传空串→&""，传值→&id）：
	//   wga: nil=不改 / &""=清空（切回默认 Supervisor）/ &子智能体id=换子智能体
	//   dip: nil=不改 / &员工id=换员工（dip 不支持清空，&"" 视为不改）
	AgentId *string           `json:"agentId"`
	Config  map[string]string `json:"config"`
}

func (r *UpdateChannelRequest) Check() error { return nil }

// UpdateChannelStatusRequest 启用/停用通道请求
type UpdateChannelStatusRequest struct {
	Enabled bool `json:"enabled" binding:"required"`
}

func (r *UpdateChannelStatusRequest) Check() error { return nil }

// ListChannelsRequest 获取通道列表请求
type ListChannelsRequest struct {
}

func (r *ListChannelsRequest) Check() error { return nil }

// GetChannelRequest 获取通道详情请求
type GetChannelRequest struct {
	ID string `uri:"id" json:"id" binding:"required"`
}

func (r *GetChannelRequest) Check() error { return nil }

// DeleteChannelRequest 删除通道请求
type DeleteChannelRequest struct {
	ID string `uri:"id" json:"id" binding:"required"`
}

func (r *DeleteChannelRequest) Check() error { return nil }

// DisconnectChannelRequest 断开通道请求
type DisconnectChannelRequest struct {
	ID string `uri:"id" json:"id" binding:"required"`
}

func (r *DisconnectChannelRequest) Check() error { return nil }

// SendMessageRequest 内部服务发消息请求（callback 接口，无鉴权）。
// channelId 必填；userId 可选，缺省时由 channel-service 自动取该通道最近互动过的 IM 用户作为收件人。
// msgType 可选：text（默认，纯文本，content 必填）/ markdown（钉钉渲染 md 卡片，微信降级纯文本，content 必填，title 可选）/
//   file（发送文件附件，fileUrl+fileName 必填，content 可空作附带文案）。
// fileUrl 为万悟 minio 文件下载地址（先调 /callback/v1/file/upload/base64 上传取得），
// channel-service 下载字节后投递，不占请求体。钉钉/微信支持文件，飞书不支持会返回错误。
type SendMessageRequest struct {
	ChannelID    string `json:"channelId" binding:"required"`
	Content      string `json:"content"`
	UserID       string `json:"userId"`
	MsgType      string `json:"msgType"`
	Title        string `json:"title"`
	FileUrl      string `json:"fileUrl"`
	FileName     string `json:"fileName"`
	FileMimeType string `json:"fileMimeType"`
}

func (r *SendMessageRequest) Check() error { return nil }
