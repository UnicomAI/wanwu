package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

// OpenAPIModelListItem OpenAPI 模型列表仅返回字段,只用于openapi
type OpenAPIModelListItem struct {
	UUID        string `json:"uuid"`
	DisplayName string `json:"displayName"`
	Provider    string `json:"provider"`
	ModelType   string `json:"modelType"`
	Model       string `json:"model"`
	ScopeType   string `json:"scopeType"`
}

type ModelInfo struct {
	ModelId      string                  `json:"modelId"`
	Uuid         string                  `json:"uuid"`
	Provider     string                  `json:"provider" validate:"required" enums:"OpenAI-API-compatible,YuanJing"` // 模型供应商
	Model        string                  `json:"model" validate:"required"`                                           // 模型名称
	ModelType    string                  `json:"modelType" validate:"required" enums:"llm,embedding,rerank"`
	DisplayName  string                  `json:"displayName"` // 模型显示名称
	Avatar       request.Avatar          `json:"avatar" `     // 模型图标路径
	PublishDate  string                  `json:"publishDate"` // 模型发布时间
	IsActive     bool                    `json:"isActive"`    // 启用状态（true: 启用，false: 禁用）
	UserId       string                  `json:"userId"`
	OrgId        string                  `json:"orgId"`
	CreatedAt    string                  `json:"createdAt"`
	UpdatedAt    string                  `json:"updatedAt"`
	ModelDesc    string                  `json:"modelDesc"`
	Tags         []mp_common.Tag         `json:"tags"`
	Config       interface{}             `json:"config"`
	Examples     *mp.ProviderModelConfig `json:"examples,omitempty"` // 仅用于swagger展示；模型对应供应商中的对应llm、embedding或rerank结构是config实际的参数
	ScopeType    string                  `json:"scopeType"`          // 模型公开范围(1-私有 2-公开 3-组织)
	AllowEdit    bool                    `json:"allowEdit"`          // 是否允许前端编辑（true: 允许编辑，false: 隐藏编辑按钮）
	ImportSource string                  `json:"importSource"`       // 模型导入来源(builtin=平台内置,external=外部URL)
}

type ModelTagItem struct {
	Text string `json:"text"`
}

type RecommendModel struct {
	Model           string         `json:"model"`
	DisplayName     string         `json:"displayName"`
	Tags            []ModelTagItem `json:"tags"`
	VisionSupport   string         `json:"visionSupport"`
	FunctionCalling string         `json:"functionCalling"`
	ThinkingSupport string         `json:"thinkingSupport"`
}

type ModelTypeItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type ProviderModelTypes struct {
	Key      string           `json:"key"`
	Name     string           `json:"name"`
	Children []*ModelTypeItem `json:"children"`
}
