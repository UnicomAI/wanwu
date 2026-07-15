package callback

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/callback"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func Register(callbackAPI *gin.RouterGroup) {
	// callback
	mid.Sub("callback").Reg(callbackAPI, "/file/url/base64", http.MethodPost, callback.FileUrlConvertBase64, "文件URL转换为base64")
	mid.Sub("callback").Reg(callbackAPI, "/file/upload/base64", http.MethodPost, callback.UploadFileByBase64, "通过base64上传文件")
	mid.Sub("callback").Reg(callbackAPI, "/file/unarchive", http.MethodPost, callback.UnarchiveFile, "解压压缩包")
	// user
	mid.Sub("callback").Reg(callbackAPI, "/user/list", http.MethodPost, callback.GetUserListByUserIds, "根据userId获取用户信息")
	// model
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId", http.MethodGet, callback.GetModelById, "根据modelId获取模型")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/chat/completions", http.MethodPost, callback.ModelChatCompletions, "Model Chat Completions")
	// 带 trace 上下文的模型代理路由（sandbox 内 opencode/eino-agent 通过 URL 路径传递 trace 信息）
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/trace/:traceId/span/:spanId/chat/completions", http.MethodPost, callback.ModelChatCompletionsWithTrace, "Model Chat Completions with Trace")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/embeddings", http.MethodPost, callback.ModelEmbeddings, "Model Embeddings")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/multimodal-embeddings", http.MethodPost, callback.ModelMultiModalEmbeddings, "Model multimodal-Embeddings")

	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/rerank", http.MethodPost, callback.ModelTextRerank, "Model rerank")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/multimodal-rerank", http.MethodPost, callback.ModelMultiModalRerank, "Model multimodal-rerank")

	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/ocr", http.MethodPost, callback.ModelOcr, "Model ocr")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/gui", http.MethodPost, callback.ModelGui, "Model gui")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/pdf-parser", http.MethodPost, callback.ModelPdfParser, "Model pdf文档解析")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/asr", http.MethodPost, callback.ModelSyncAsr, "Model sync asr")
	// workflow
	mid.Sub("callback").Reg(callbackAPI, "/workflow/list", http.MethodGet, callback.GetWorkflowList, "根据userId和spaceId获取Workflow")
	mid.Sub("callback").Reg(callbackAPI, "/workflow/tool/square", http.MethodGet, callback.GetWorkflowSquareTool, "获取内置工具详情")
	mid.Sub("callback").Reg(callbackAPI, "/workflow/tool/custom", http.MethodGet, callback.GetWorkflowCustomTool, "获取自定义工具详情")
	// mcp
	mid.Sub("callback").Reg(callbackAPI, "/mcp", http.MethodGet, callback.GetMCP, "获取自定义MCP详情")
	mid.Sub("callback").Reg(callbackAPI, "/mcp/server", http.MethodGet, callback.GetMCPServer, "获取MCP服务详情")
	// chatflow
	mid.Sub("callback").Reg(callbackAPI, "/chatflow/list", http.MethodGet, callback.GetChatflowList, "根据userId和spaceId获取Chatflow")
	// agent bff proxy
	mid.Sub("callback").Reg(callbackAPI, "/agent/:assistantId/chat", http.MethodPost, callback.AgentChatProxy, "智能体代理问答")
	// rag bff proxy
	mid.Sub("callback").Reg(callbackAPI, "/rag/search-knowledge-base", http.MethodPost, callback.SearchKnowledgeBase, "查询知识库列表（命中测试）")
	mid.Sub("callback").Reg(callbackAPI, "/rag/knowledge/stream/search", http.MethodPost, callback.KnowledgeStreamSearch, "根据知识库id 和当前用户id 获取有权限的知识库列表信息")
	// rag bff proxy
	mid.Sub("callback").Reg(callbackAPI, "/rag/search-QA-base", http.MethodPost, callback.SearchQABase, "查询问答库列表（命中测试）")
	// wga rag proxy
	mid.Sub("callback").Reg(callbackAPI, "/wga/rag/search-knowledge-base", http.MethodPost, callback.WgaRagSearchKnowledgeBase, "WGA知识库检索")
	// wga sandbox
	mid.Sub("callback").Reg(callbackAPI, "/wga/sandbox/run", http.MethodPost, callback.WgaSandboxRun, "WGA沙箱运行")
	mid.Sub("callback").Reg(callbackAPI, "/wga/sandbox/cleanup", http.MethodPost, callback.WgaSandboxCleanup, "WGA沙箱清理")
	// app record
	mid.Sub("callback").Reg(callbackAPI, "/app/record", http.MethodPost, callback.AppRecord, "应用使用记录")
	// skill
	mid.Sub("callback").Reg(callbackAPI, "/skill/builtin/list", http.MethodPost, callback.SearchBuiltInSkillList, "内置skill详情列表")
	mid.Sub("callback").Reg(callbackAPI, "/skill/custom/list", http.MethodPost, callback.SearchCustomSkillList, "自定义skill详情列表")
	mid.Sub("callback").Reg(callbackAPI, "/skill/acquired/list", http.MethodPost, callback.SearchAcquiredSkillList, "我添加skill详情列表")
	mid.Sub("callback").Reg(callbackAPI, "/skill/detail", http.MethodGet, callback.GetSkillDetail, "获取技能详情")

	// --- 通道内部发消息（无鉴权，供内部服务调用）---
	mid.Sub("callback").Reg(callbackAPI, "/channel/send-message", http.MethodPost, callback.SendMessage, "内部服务发消息（无鉴权）")
}
