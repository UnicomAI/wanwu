package request

// AdminUserSelect admin 选择的用户信息
type AdminUserSelect struct {
	IsAllOrg   bool     `json:"isAllOrg"` // 是否全选组织
	OrgIdList  []string `json:"orgIdList"`
	UserIdList []string `json:"userIdList"`
}

type AdminPublish struct {
	PublishStatus []string `json:"publishStatus"` // draft:草稿，publish:发布
	PublishScope  []string `json:"publishScope"`  // private：私密发布，organization:组织内发布，public：公开发布
}

type AdminKnowledgePageListReq struct {
	Name     string  `json:"name" form:"name" `
	Category []int32 `json:"category" form:"category"`  // 0:知识库，1:问答库，2:多模态知识库
	External int32   `json:"external" form:"external" ` // -1:全部，0:内部知识库，1:外部知识库
	AdminPublish
	AdminUserSelect
	PageSearch
	CommonCheck
}

type AdminKnowledgeDetailReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId"  validate:"required"` // 知识库id
	CommonCheck
}

type AdminKnowledgeFileListReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId"  validate:"required"` // 知识库id
	PageSearch
	CommonCheck
}

type AdminKnowledgeFileDetailReq struct {
	KnowledgeId string `json:"knowledgeId" form:"knowledgeId"  validate:"required"` // 知识库id
	DocId       string `json:"docId" form:"docId"  validate:"required"`             // 文档id
	PageSearch
	CommonCheck
}

type AdminWorkflowPageListReq struct {
	Name    string   `json:"name" form:"name" `
	AppType []string `json:"appType" form:"appType"` // workflow:工作流，chatflow:对话流
	AdminPublish
	AdminUserSelect
	PageSearch
	CommonCheck
}

type AdminSkillPageListReq struct {
	Name string `json:"name" form:"name" `
	AdminPublish
	AdminUserSelect
	PageSearch
	CommonCheck
}

type AdminSkillDetailReq struct {
	SkillId   string `json:"skillId" form:"skillId" validate:"required"`
	SkillType string `json:"skillType" form:"skillType"` //builtin:内置技能，acquired:自定义技能,不填写默认自定义
	CommonCheck
}

type AdminModelPageListReq struct {
	Name      string   `json:"name" form:"name"`           // 模型名称
	ModelType []string `json:"modelType" form:"modelType"` // 模型类型：llm, rerank, embedding, multimodal-rerank, multimodal-embedding, ocr, gui, pdf-parser, sync-asr, text2image
	Provider  string   `json:"provider" form:"provider"`   // 模型供应商: OpenAI-API-compatible,YuanJing, HuoShan, Ollama, Qwen, Infini, QianFan, DeepSeek, Jina, ZhiPu
	AdminPublish
	AdminUserSelect
	PageSearch
	CommonCheck
}

type AdminModelDetailReq struct {
	ModelId string `json:"modelId" form:"modelId" validate:"required"` // 模型id
	CommonCheck
}

type AdminRagPageListReq struct {
	Name string `json:"name" form:"name"` // 知识问答名称
	AdminPublish
	AdminUserSelect
	PageSearch
	CommonCheck
}

type AdminRagDetailReq struct {
	RagId string `json:"ragId" form:"ragId" validate:"required"` // 知识问答id
	CommonCheck
}
