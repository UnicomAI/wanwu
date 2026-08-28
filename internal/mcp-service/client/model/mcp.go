package model

type MCPClient struct {
	ID            uint32 `gorm:"primary_key"`
	CreatedAt     int64  `gorm:"autoCreateTime:milli;index:idx_mcp_created_at"`
	UpdatedAt     int64  `gorm:"autoUpdateTime:milli"`
	OrgID         string `gorm:"index:idx_mcp_org_id"`
	UserID        string `gorm:"index:idx_mcp_user_id"`
	McpSquareId   string `gorm:"index:idx_mcp_mcp_square_id"`
	Name          string `gorm:"index:idx_mcp_name"`
	From          string `gorm:"index:idx_mcp_from"`
	Desc          string
	SseUrl        string
	StreamableUrl string
	Transport     string // 传输协议: "sse" 或 "streamable"
	AvatarPath    string `gorm:"column:avatar_path;comment:'自定义工具头像'"`
	AuthJSON      string `gorm:"column:auth_json;type:longtext;comment:'鉴权json'"`
	Headers       string `gorm:"column:headers;type:longtext;comment:'请求头json'"`
}
