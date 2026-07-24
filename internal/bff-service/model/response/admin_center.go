package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type OwnerInfo struct {
	OwnerUserId   string `json:"ownerUserId"`   // 拥有者用户id（知识库可转让，未必是创建者）
	OwnerUserName string `json:"ownerUserName"` // 拥有者用户名称（知识库可转让，未必是创建者）
	OwnerOrgId    string `json:"ownerOrgId"`    // 拥有者组织id
	OwnerOrgName  string `json:"ownerOrgName"`  // 拥有者组织id
}

type OwnerInfoService interface {
	GetOwnerInfo() OwnerInfo
	SetOwnerInfo(OwnerInfo)
}

// OwnerHolder 内嵌 OwnerInfo，并实现读写接口
type OwnerHolder struct {
	OwnerInfo
}

func (h *OwnerHolder) GetOwnerInfo() OwnerInfo {
	return h.OwnerInfo
}

func (h *OwnerHolder) SetOwnerInfo(info OwnerInfo) {
	h.OwnerInfo = info
}

func CreateOwnerHolder(ownerUserId, ownerOrgId string) OwnerHolder {
	return OwnerHolder{OwnerInfo: OwnerInfo{
		OwnerUserId: ownerUserId,
		OwnerOrgId:  ownerOrgId,
	}}
}

type AuthorizedPersonnel struct {
	UserId  string         `json:"userId"`
	Name    string         `json:"name"`
	OrgId   string         `json:"orgId"`
	OrgName string         `json:"orgName"`
	Avatar  request.Avatar `json:"avatar"`
}

type AdminAppBaseInfo struct {
	Avatar                  request.Avatar        `json:"avatar"`                  // 应用图标
	Name                    string                `json:"name"`                    // 应用名称
	Desc                    string                `json:"desc"`                    // 应用描述
	CreatedAt               string                `json:"createdAt"`               // 应用创建时间
	UpdatedAt               string                `json:"updatedAt"`               // 应用更新时间(用于历史记录排序)
	PublishStatus           string                `json:"publishStatus"`           // draft:草稿，publish:发布
	PublishScope            string                `json:"publishScope"`            // private：私密发布，organization:组织内发布，authorized_personnel:指定人员发布 public：公开发布
	AuthorizedPersonnelList []AuthorizedPersonnel `json:"authorizedPersonnelList"` // 指定人员列表
	OwnerHolder
}

type AdminKnowledge struct {
	KnowledgeId string         `json:"knowledgeId"` // 知识库id
	Name        string         `json:"name"`        // 知识库名称
	Description string         `json:"description"` // 知识库描述
	Category    int32          `json:"category"`    // 0:知识库 1:问答库 2:多模态知识库
	Avatar      request.Avatar `json:"avatar"`      // 头像
	External    int32          `json:"external"`    // 0:内部知识库 1:外部知识库
	CreatedAt   string         `json:"createdAt"`   // 创建时间
	UpdatedAt   string         `json:"updatedAt"`   // 更新时间
	OwnerHolder
}

type AdminKnowledgeBase struct {
	KnowledgeId    string          `json:"knowledgeId"`
	Name           string          `json:"name"`
	GraphSwitch    int32           `json:"graphSwitch"`
	Description    string          `json:"description"`
	Keywords       []*KeywordsInfo `json:"keywords"`
	EmbeddingModel *ModelInfo      `json:"embeddingModel"`
	LlmModelId     string          `json:"llmModelId"`
	Category       int32           `json:"category"` // 0: 知识库 1: 问答库 2: 多模态知识库
	Avatar         request.Avatar  `json:"avatar"`   // 头像
	AdminAppBaseInfo
}

type AdminWorkflow struct {
	AppBriefInfo
	OwnerHolder
}

type AdminSkillDetail struct {
	PublishedSkillInfo
	OwnerHolder
}

type AdminModel struct {
	ModelInfo
	OwnerHolder
}

type AdminModelBase struct {
	AdminAppBaseInfo
	Provider string `json:"provider"` // 模型供应商
}

type AdminRag struct {
	AppBriefInfo
	OwnerHolder
}

type AdminRagBase struct {
	AdminAppBaseInfo
}

type AdminRagDetail struct {
	RagInfo
	LlmModel      *ModelInfo `json:"llmModel"`      // llm模型（含头像/标签）
	RerankModel   *ModelInfo `json:"rerankModel"`   // 知识库Rerank模型（含头像/标签）
	QaRerankModel *ModelInfo `json:"qaRerankModel"` // 问答库Rerank模型（含头像/标签）
}

type AdminMCP struct {
	Avatar      request.Avatar `json:"avatar"`      // logo
	MCPID       string         `json:"mcpId"`       // mcpId
	Type        string         `json:"type"`        // 类型
	Name        string         `json:"name"`        // 名称
	Description string         `json:"description"` // 描述
	ServerFrom  string         `json:"serverFrom"`  // 来源
	UpdatedAt   string         `json:"updatedAt"`   // 更新时间
	OwnerHolder
}

type AdminMCPBase struct {
	AdminAppBaseInfo
	Type string `json:"type"` // 类型
}

type AdminTool struct {
	Avatar      request.Avatar `json:"avatar"`      // logo
	ToolID      string         `json:"toolId"`      // toolId
	Name        string         `json:"name"`        // 名称
	Description string         `json:"description"` // 描述
	UpdatedAt   string         `json:"updatedAt"`   // 更新时间
	OwnerHolder
}

type AdminToolBase struct {
	AdminAppBaseInfo
}

type AdminPrompt struct {
	Avatar      request.Avatar `json:"avatar"`      // logo
	PromptID    string         `json:"promptId"`    // promptId
	Name        string         `json:"name"`        // 名称
	Description string         `json:"description"` // 描述
	UpdatedAt   string         `json:"updatedAt"`   // 更新时间
	OwnerHolder
}

type AdminPromptBase struct {
	AdminAppBaseInfo
}

type AdminAssistantBase struct {
	AdminAppBaseInfo
	Category      int32 `json:"category"`      // 智能体分类(1:单智能体,2:多智能体)
	HideKnowledge int32 `json:"hideKnowledge"` // 是否隐藏知识库引用(1:隐藏,0:显示)，仅已发布版本有值
}

type AdminAssistant struct {
	AppBriefInfo
	OwnerHolder
}

type AdminAssistantDetail struct {
	Assistant
	OwnerHolder
}

type AdminSensitiveWord struct {
	SensitiveWordTableDetail
	OwnerHolder
}

type AdminSensitiveWordBase struct {
	TableId   string `json:"tableId"`
	TableName string `json:"tableName"`
	Remark    string `json:"remark"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	OwnerHolder
}

type AdminSensitiveWordDetailResp struct {
	Reply    string                           `json:"reply"`
	List     []*SensitiveWordVocabularyDetail `json:"list"`
	Total    int64                            `json:"total"`
	PageNo   int                              `json:"pageNo"`
	PageSize int                              `json:"pageSize"`
}
