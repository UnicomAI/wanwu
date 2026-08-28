package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerWorkflow(apiV1 *gin.RouterGroup) {
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/workflow/list", http.MethodGet, v1.GetWorkflowAndChatflowList, "获取工作流/对话流应用列表")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/workflow", http.MethodPost, v1.CreateWorkflow, "创建工作流或对话流")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/workflow/convert", http.MethodPost, v1.WorkflowConvert, "workflow转为chatflow")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/workflow/copy/draft", http.MethodPost, v1.CopyWorkflowDraft, "拷贝workflow草稿")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/workflow/model/select/llm", http.MethodGet, v1.ListLlmModelsByWorkflow, "llm模型列表（用于workflow）")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/workflow/export", http.MethodGet, v1.ExportWorkflow, "导出workflow")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/workflow/export/draft", http.MethodGet, v1.ExportWorkflowDraft, "导出workflow草稿")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/workflow/import", http.MethodPost, v1.ImportWorkflow, "导入workflow")

	mid.Sub("app.workflow").Reg(apiV1, "/workflow/tool/action", http.MethodGet, v1.GetWorkflowToolDetail, "获取Tool具体action")
	mid.Sub("app.workflow").Reg(apiV1, "/workflow/tool/select", http.MethodGet, v1.GetWorkflowToolSelect, "获取Tool列表")
	mid.Sub("app.workflow").Reg(apiV1, "/workflow/tool/box", http.MethodGet, v1.GetWorkflowToolBoxDetail, "工具箱明细查询（只用于适配本体智能体）")

	mid.Sub("app.workflow").Reg(apiV1, "/workflow/select", http.MethodGet, v1.GetWorkflowSelect, "智能体工作流下拉列表接口")

	mid.Sub("app.workflow").Reg(apiV1, "/workflow/template", http.MethodPost, v1.CreateWorkflowByTemplate, "复制工作流模板")
	// --- chatflow ---
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/chatflow/convert", http.MethodPost, v1.ChatflowConvert, "chatflow转为workflow")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/chatflow/copy/draft", http.MethodPost, v1.CopyChatflowDraft, "拷贝chatflow草稿")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/chatflow/import", http.MethodPost, v1.ImportChatflow, "导入chatflow")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/chatflow/export", http.MethodGet, v1.ExportChatflow, "导出chatflow")
	mid.Sub("app.workflow").Reg(apiV1, "/appspace/chatflow/export/draft", http.MethodGet, v1.ExportChatflowDraft, "导出chatflow草稿")

}
