package request

type RagBrief struct {
	RagID string `json:"ragId" validate:"required"`
	AppBriefConfig
}

type RagConfig struct {
	RagID               string                 `json:"ragId" validate:"required"`
	ModelConfig         AppModelConfig         `json:"modelConfig" validate:"required"`         // Model
	RerankConfig        AppModelConfig         `json:"rerankConfig" validate:"required"`        // Rerank model
	KnowledgeBaseConfig AppKnowledgebaseConfig `json:"knowledgeBaseConfig" validate:"required"` // knowledge base
	SafetyConfig        AppSafetyConfig        `json:"safetyConfig"`                            // Sensitive word list configuration
}

type ChatRagRequest struct {
	RagID    string     `json:"ragId" validate:"required"`
	Question string     `json:"question" validate:"required"`
	History  []*History `json:"history"`
}

type History struct {
	Query       string `json:"query"`
	Response    string `json:"response"`
	NeedHistory bool   `json:"needHistory"`
}

type RagReq struct {
	RagID string `form:"ragId" json:"ragId" validate:"required"`
}

func (r RagBrief) Check() error {
	return nil
}

func (r RagConfig) Check() error {
	if err := r.ModelConfig.Check(); err != nil {
		return err
	}
	if err := r.RerankConfig.Check(); err != nil {
		return err
	}
	return nil
}

func (c ChatRagRequest) Check() error {
	return nil
}

func (r RagReq) Check() error {
	return nil
}
