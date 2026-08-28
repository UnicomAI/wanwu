package config

import (
	"github.com/UnicomAI/wanwu/api/proto/common"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
)

type AssistantConversionHistory struct {
	Query         string   `json:"query"`
	UploadFileUrl []string `json:"upload_file_url,omitempty"`
	Response      string   `json:"response"`
}

type KnParams struct {
	KnowledgeBase        []string               `json:"knowledgeBase"`   // 知识库名称列表
	KnowledgeIdList      []string               `json:"knowledgeIdList"` // 知识库id列表
	RerankId             interface{}            `json:"rerank_id"`
	Model                interface{}            `json:"model"`
	ModelUrl             interface{}            `json:"model_url"`
	RerankMod            string                 `json:"rerank_mod"`
	RetrieveMethod       string                 `json:"retrieve_method"`
	Weights              *WeightParams          `json:"weights,omitempty"`
	MaxHistory           int32                  `json:"max_history"`
	Threshold            float32                `json:"threshold"`
	TopK                 int32                  `json:"topK"`
	RewriteQuery         bool                   `json:"rewrite_query"`
	TermWeight           float32                `json:"term_weight_coefficient"`       // 关键词系数, 默认为1
	MetaFilter           bool                   `json:"metadata_filtering"`            // 元数据过滤开关
	MetaFilterConditions []*MetadataFilterParam `json:"metadata_filtering_conditions"` // 元数据过滤条件
	UseGraph             bool                   `json:"use_graph"`                     // 知识图谱开关
}

type MetadataFilterParam struct {
	FilterKnowledgeName string                `json:"filtering_kb_name"`
	LogicalOperator     string                `json:"logical_operator"`
	MetaList            []*MetadataFilterItem `json:"conditions"` // 元数据过滤列表
}

type MetadataFilterItem struct {
	MetaName           string      `json:"meta_name"`           // 元数据名称
	MetaType           string      `json:"meta_type"`           // 元数据类型
	ComparisonOperator string      `json:"comparison_operator"` // 比较运算符
	Value              interface{} `json:"value,omitempty"`     // 用于过滤的条件值
}

type WeightParams struct {
	VectorWeight float32 `json:"vector_weight"` //语义权重
	TextWeight   float32 `json:"text_weight"`   //关键字权重
}

type AgentSSERequest struct {
	Input            string                       `json:"input"`
	Stream           bool                         `json:"stream"`
	SystemRole       string                       `json:"system_role,omitempty"`
	UploadFileUrl    []string                     `json:"upload_file_url,omitempty"`
	FileName         string                       `json:"file_name,omitempty"`
	PluginList       []PluginListAlgRequest       `json:"plugin_list,omitempty"`
	Model            string                       `json:"model,omitempty"`
	ModelUrl         string                       `json:"model_url,omitempty"`
	SearchUrl        string                       `json:"search_url,omitempty"`
	SearchKey        string                       `json:"search_key,omitempty"`
	SearchRerankId   interface{}                  `json:"search_rerank_id,omitempty"`
	UseSearch        bool                         `json:"use_search,omitempty"`
	KnParams         *KnParams                    `json:"kn_params,omitempty"`
	UseKnow          bool                         `json:"use_know,omitempty"`
	ModelId          string                       `json:"model_id,omitempty"`
	History          []AssistantConversionHistory `json:"history,omitempty"`
	McpTools         map[string]MCPToolInfo       `json:"mcp_tools,omitempty"`
	ToolsName        []string                     `json:"tools_name,omitempty"`
	AutoCitation     bool                         `json:"auto_citation,omitempty"`
	MaxHistoryLength int32                        `json:"max_history_length,omitempty"`
	ModelParams      map[string]interface{}       `json:"-"` // 用于合并动态模型参数，不直接序列化到JSON
}

type PluginListAlgRequest struct {
	APISchema map[string]interface{} `json:"api_schema"`
	APIAuth   *openapi3_util.Auth    `json:"api_auth,omitempty"`
}

type MCPToolInfo struct {
	URL          string                    `json:"url"`
	Transport    string                    `json:"transport"`
	ToolNameList []string                  `json:"toolNameList"`
	Avatar       string                    `json:"avatar"`
	Headers      map[string]string         `json:"headers"`
	ApiAuth      *common.ApiAuthWebRequest `json:"apiAuth"`
}

type ToolsMap map[string]MCPToolInfo

type RequestData struct {
	McpTools ToolsMap `json:"mcp_tools"`
}
