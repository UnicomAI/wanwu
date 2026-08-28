package response

type ToolBoxDetail struct {
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
	HasNext    bool              `json:"has_next"`
	HasPrev    bool              `json:"has_prev"`
	BoxID      string            `json:"box_id"`
	APIKey     string            `json:"api_key"`  // 工具箱共享的 apiKey 值
	APIAuth    ToolBoxAPIAuth    `json:"api_auth"` // 工具箱共享的鉴权配置
	Tools      []ToolBoxToolItem `json:"tools"`
}

type ToolBoxAPIAuth struct {
	AuthType           string `json:"auth_type"`
	APIKeyHeaderPrefix string `json:"api_key_header_prefix"`
	APIKeyHeader       string `json:"api_key_header"`
	APIKeyQueryParam   string `json:"api_key_query_param"`
	APIKeyValue        string `json:"api_key_value"`
}

type ToolBoxToolItem struct {
	ToolID           string              `json:"tool_id"`
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	Status           string              `json:"status"`
	MetadataType     string              `json:"metadata_type"`
	Metadata         ToolBoxMetadata     `json:"metadata"`
	UseRule          string              `json:"use_rule"`
	GlobalParameters ToolBoxGlobalParams `json:"global_parameters"`
	CreateTime       int64               `json:"create_time"`
	UpdateTime       int64               `json:"update_time"`
	CreateUser       string              `json:"create_user"`
	UpdateUser       string              `json:"update_user"`
	ExtendInfo       map[string]any      `json:"extend_info"`
	ResourceObject   string              `json:"resource_object"`
}

type ToolBoxMetadata struct {
	Version     string         `json:"version"`
	Summary     string         `json:"summary"`
	Description string         `json:"description"`
	ServerURL   string         `json:"server_url"`
	Path        string         `json:"path"`
	Method      string         `json:"method"`
	CreateTime  int64          `json:"create_time"`
	UpdateTime  int64          `json:"update_time"`
	CreateUser  string         `json:"create_user"`
	UpdateUser  string         `json:"update_user"`
	APISpec     ToolBoxAPISpec `json:"api_spec"`
}

type ToolBoxAPISpec struct {
	Parameters   []any                 `json:"parameters"`
	RequestBody  any                   `json:"request_body"`
	Responses    []ToolBoxResponseItem `json:"responses"`
	Components   any                   `json:"components"`
	Callbacks    any                   `json:"callbacks"`
	Security     any                   `json:"security"`
	Tags         []string              `json:"tags"`
	ExternalDocs any                   `json:"external_docs"`
}

type ToolBoxResponseItem struct {
	StatusCode  string `json:"status_code"`
	Description string `json:"description"`
	Content     any    `json:"content"`
}

type ToolBoxGlobalParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	In          string `json:"in"`
	Type        string `json:"type"`
	Value       any    `json:"value"`
}
