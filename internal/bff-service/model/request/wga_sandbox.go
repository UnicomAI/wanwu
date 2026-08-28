package request

import "github.com/UnicomAI/wanwu/pkg/util"

type WgaSandboxRunReq struct {
	ThreadID       string              `json:"threadId"`
	RunID          string              `json:"runId"`
	Model          AppModelConfig      `json:"model" validate:"required"`
	Instruction    string              `json:"instruction"`
	OverallTask    string              `json:"overallTask"`
	Messages       []WgaSandboxMessage `json:"messages"`
	Tools          []WgaSandboxTool    `json:"tools"`
	Skills         []WgaSandboxSkill   `json:"skills"`
	MCPs           []WgaSandboxMCP     `json:"mcps"`
	InputDir       string              `json:"inputDir"`
	OutputDir      string              `json:"outputDir"`
	EnableThinking bool                `json:"enableThinking"`
	SkipCleanup    bool                `json:"skipCleanup"`
	AgentName      string              `json:"agentName"`
}

type WgaSandboxMessage struct {
	Role    string `json:"role" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type WgaSandboxTool struct {
	Schema       string                  `json:"schema" validate:"required"` // OpenAPI 3.0 schema JSON 字符串
	OperationIDs []string                `json:"operationIds"`               // 允许的 operations，空=全部允许
	ApiAuth      *util.ApiAuthWebRequest `json:"apiAuth"`                    // API 认证配置
}

type WgaSandboxSkillVariable struct {
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description"`
	VariableKey   string `json:"variableKey" validate:"required"`
	VariableValue string `json:"variableValue" validate:"required"`
}

type WgaSandboxSkill struct {
	Dir       string                    `json:"dir" validate:"required"`
	Variables []WgaSandboxSkillVariable `json:"variables"`
}

type WgaSandboxMCP struct {
	Name        string `json:"name" validate:"required"`
	URL         string `json:"url" validate:"required"`
	Description string `json:"description"`
}

func (r *WgaSandboxRunReq) Check() error {
	return nil
}

type WgaSandboxCleanupReq struct {
	RunID string `json:"runId" validate:"required"`
}

func (r *WgaSandboxCleanupReq) Check() error {
	return nil
}
