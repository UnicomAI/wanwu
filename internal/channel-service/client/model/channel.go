package model

import "github.com/UnicomAI/wanwu/pkg/util"

// Channel 通道数据模型
type Channel struct {
	ID          uint32 `gorm:"primary_key"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli;index:idx_channel_created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli"`
	ChannelID   string `gorm:"uniqueIndex:idx_channel_channel_id;size:64"`
	Name        string `gorm:"index:idx_channel_name;size:128"`
	ChannelType string `gorm:"index:idx_channel_type;size:32"`   // wechat/dingtalk/feishu
	Status      string `gorm:"index:idx_channel_status;size:32"` // loggedIn/offline
	Enabled     bool   `gorm:"index:idx_channel_enabled"`
	AppType     string `gorm:"size:32"`                          // agent
	AppID       string `gorm:"index:idx_channel_app_id;size:64"` // 智能体 UUID（agent）/ 子智能体id（wga）/ 数字员工id（dip）
	AppName     string `gorm:"size:128"`                         // 智能体名称（冗余展示）：agent=智能体名 / wga=子智能体名或"通用智能体" / dip=数字员工名
	ApiKeyID    string `gorm:"size:64"`                          // API Key ID
	ApiKeyName  string `gorm:"size:128"`                         // API Key 名称（冗余展示）
	ApiKey      string `gorm:"size:512"`                         // API Key（明文存储）
	ModelUuid   string `gorm:"size:64"`                          // WGA 通道使用的模型 UUID
	AgentId     string `gorm:"size:64"`                          // wga=绑定的子智能体id（直连该子智能体，跳过 Supervisor）/ dip=数字员工id
	Config      string `gorm:"type:text"`                        // 平台配置 JSON
	AccountId   string `gorm:"size:64"`                          // 平台账号 ID
	Nickname    string `gorm:"size:128"`
	Avatar      string `gorm:"size:512"`
	UserID      string `gorm:"index:idx_channel_user_id;size:64"`
	OrgID       string `gorm:"index:idx_channel_org_id;size:64"`
}

// TableName 表名
func (Channel) TableName() string {
	return "channels"
}

// HasApiKey 是否已绑定 API Key
func (c *Channel) HasApiKey() bool {
	return c.ApiKey != ""
}

// QRSession 扫码登录会话
type QRSession struct {
	ID          uint32 `gorm:"primary_key"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli"`
	SessionID   string `gorm:"uniqueIndex:idx_qr_session_id;size:128"`
	ChannelType string `gorm:"index:idx_qr_session_type;size:32"`   // wechat/dingtalk/feishu
	Status      string `gorm:"index:idx_qr_session_status;size:32"` // waiting/scanned/confirmed/success/expired/error
	Credentials string `gorm:"type:text"`                           // 加密的平台凭据 JSON
	DeviceCode  string `gorm:"size:256"`                            // 钉钉/飞书 Device Flow 的 device_code
	UserID      string `gorm:"size:64"`
	OrgID       string `gorm:"size:64"`
	ExpireAt    int64  `gorm:"index:idx_qr_session_expire_at"`
}

// TableName 表名
func (QRSession) TableName() string {
	return "qr_sessions"
}

// NewChannelID 生成通道 ID
func NewChannelID() string {
	return "ch_" + util.GenUUID()
}

// NewSessionID 生成扫码会话 ID
func NewSessionID() string {
	return "qr_" + util.GenUUID()
}
