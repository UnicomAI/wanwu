package config

import (
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
)

type AgentScopeWorkFlowSchemaResp struct {
	Code int `json:"code"`
	Data struct {
		Base64OpenAPISchema string `json:"base64OpenAPISchema"`
	} `json:"data"`
}

type AssistantConversionHistory struct {
	Query         string   `json:"query"`
	UploadFileUrl []string `json:"upload_file_url,omitempty"`
	Response      string   `json:"response"`
}

type KnParams struct {
	KnowledgeBase        []string               `json:"knowledgeBase"`   // 知识库名称列表 [EN] Knowledge base name list
	KnowledgeIdList      []string               `json:"knowledgeIdList"` // 知识库id列表 [EN] Knowledge base id list
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
	TermWeight           float32                `json:"term_weight_coefficient"`       // 关键词系数, 默认为1 [EN] Keyword coefficient, default is 1
	MetaFilter           bool                   `json:"metadata_filtering"`            // 元数据过滤开关 [EN] Metadata filter switch
	MetaFilterConditions []*MetadataFilterParam `json:"metadata_filtering_conditions"` // 元数据过滤条件 [EN] Metadata filters
	UseGraph             bool                   `json:"use_graph"`                     // 知识图谱开关 [EN] Knowledge graph switch
}

type MetadataFilterParam struct {
	FilterKnowledgeName string                `json:"filtering_kb_name"`
	LogicalOperator     string                `json:"logical_operator"`
	MetaList            []*MetadataFilterItem `json:"conditions"` // 元数据过滤列表 [EN] Metadata filter list
}

type MetadataFilterItem struct {
	MetaName           string      `json:"meta_name"`           // 元数据名称 [EN] metadata name
	MetaType           string      `json:"meta_type"`           // 元数据类型 [EN] metadata type
	ComparisonOperator string      `json:"comparison_operator"` // 比较运算符 [EN] comparison operator
	Value              interface{} `json:"value,omitempty"`     // 用于过滤的条件值 [EN] condition value for filtering
}

type WeightParams struct {
	VectorWeight float32 `json:"vector_weight"` //语义权重 [EN] semantic weight
	TextWeight   float32 `json:"text_weight"`   //关键字权重 [EN] keyword weight
}

type AgentSSERequest struct {
	Input          string                       `json:"input"`
	Stream         bool                         `json:"stream"`
	SystemRole     string                       `json:"system_role,omitempty"`
	UploadFileUrl  []string                     `json:"upload_file_url,omitempty"`
	FileName       string                       `json:"file_name,omitempty"`
	PluginList     []PluginListAlgRequest       `json:"plugin_list,omitempty"`
	Model          string                       `json:"model,omitempty"`
	ModelUrl       string                       `json:"model_url,omitempty"`
	SearchUrl      string                       `json:"search_url,omitempty"`
	SearchKey      string                       `json:"search_key,omitempty"`
	SearchRerankId interface{}                  `json:"search_rerank_id,omitempty"`
	UseSearch      bool                         `json:"use_search,omitempty"`
	KnParams       *KnParams                    `json:"kn_params,omitempty"`
	UseKnow        bool                         `json:"use_know,omitempty"`
	ModelId        string                       `json:"model_id,omitempty"`
	History        []AssistantConversionHistory `json:"history,omitempty"`
	McpTools       map[string]MCPToolInfo       `json:"mcp_tools,omitempty"`
	ToolsName      []string                     `json:"tools_name,omitempty"`
	AutoCitation   bool                         `json:"auto_citation,omitempty"`
	ModelParams    map[string]interface{}       `json:"-"` // 用于合并动态模型参数，不直接序列化到JSON [EN] Used to merge dynamic model parameters, not directly serialized to JSON
}

type PluginListAlgRequest struct {
	APISchema map[string]interface{} `json:"api_schema"`
	APIAuth   *openapi3_util.Auth    `json:"api_auth,omitempty"`
}

type MCPToolInfo struct {
	URL       string `json:"url"`
	Transport string `json:"transport"`
}

type ToolsMap map[string]MCPToolInfo

type RequestData struct {
	McpTools ToolsMap `json:"mcp_tools"`
}
