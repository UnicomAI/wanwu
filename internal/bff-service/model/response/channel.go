package response

// ChannelResponse 通道响应
type ChannelResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ChannelType string            `json:"channelType"`
	Status      string            `json:"status"`
	AccountId   string            `json:"accountId"`
	Nickname    string            `json:"nickname"`
	Avatar      string            `json:"avatar"`
	Enabled     bool              `json:"enabled"`
	AppType     string            `json:"appType"`
	AppId       string            `json:"appId"`
	AppName     string            `json:"appName"`
	ApiKeyId    string            `json:"apiKeyId"`
	ApiKeyName  string            `json:"apiKeyName"`
	HasApiKey   bool              `json:"hasApiKey"`
	ModelUuid   string            `json:"modelUuid"`
	AgentId     string            `json:"agentId"` // WGA 通道绑定的子智能体 ID（直连该子智能体，跳过 Supervisor）
	Config      map[string]string `json:"config"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

// QRLoginResponse 扫码登录响应
type QRLoginResponse struct {
	SessionID  string `json:"sessionId"`
	QrUrl      string `json:"qrUrl"`
	ExpireAt   int64  `json:"expireAt"`
	ExpireTime int64  `json:"expireTime"`
}

// QRLoginStatusResponse 扫码状态响应
type QRLoginStatusResponse struct {
	Status      string            `json:"status"`
	Credentials map[string]string `json:"credentials,omitempty"`
	Error       string            `json:"error,omitempty"`
	BaseUrl     string            `json:"baseUrl,omitempty"` // 微信扫码成功时返回的平台 baseUrl
}

// DisconnectChannelResponse 断开通道响应
type DisconnectChannelResponse struct {
	Message string `json:"message"`
}

// WanwuApiKeyResponse 万悟 API Key 信息（用于通道选择下拉）
type WanwuApiKeyResponse struct {
	KeyID string `json:"keyId"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
}

// WanwuAgentResponse 万悟智能体信息（用于通道选择下拉）
type WanwuAgentResponse struct {
	AppId   string `json:"appId"`
	AppType string `json:"appType"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Avatar  string `json:"avatar"`
}
