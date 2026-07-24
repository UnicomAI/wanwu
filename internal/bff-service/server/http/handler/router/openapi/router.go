package openapi

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/openapi"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func Register(openAPI *gin.RouterGroup) {
	// agent — 基础管理
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent", http.MethodPost, openapi.CreateAgent, "创建智能体OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent", http.MethodPut, openapi.UpdateAgent, "更新智能体基本信息OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent", http.MethodDelete, openapi.DeleteAgent, "删除智能体OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/list", http.MethodGet, openapi.ListAgents, "智能体列表OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/info", http.MethodGet, openapi.GetAgentInfo, "获取智能体详情OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/config", http.MethodPut, openapi.UpdateAgentConfig, "更新智能体配置OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthModelByUuid([]string{"modelConfig.modelId", "rerankConfig.modelId", "recommendConfig.modelConfig.modelId"}))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/publish", http.MethodPost, openapi.PublishAgent, "发布智能体OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// agent — 已发布对话管理
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/conversation", http.MethodPost, openapi.CreateAgentConversation, "智能体创建对话OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/conversation", http.MethodDelete, openapi.DeleteAgentConversation, "删除智能体对话OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/conversation/list", http.MethodGet, openapi.ListAgentConversations, "智能体对话列表OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/conversation/detail", http.MethodGet, openapi.GetAgentConversationDetail, "智能体对话历史消息OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/conversation/clear", http.MethodDelete, openapi.ClearAgentConversation, "清空智能体对话历史OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// agent — 草稿态对话管理
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/conversation/draft/detail", http.MethodGet, openapi.GetAgentDraftConversationDetail, "草稿智能体对话历史消息OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/conversation/draft", http.MethodDelete, openapi.DeleteAgentDraftConversation, "删除草稿智能体对话历史OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// agent — 问答
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/chat", http.MethodPost, openapi.ChatAgent, "智能体问答OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordFromReq))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/agent/chat/draft", http.MethodPost, openapi.DraftChatAgent, "智能体草稿态对话OpenAPI", constant.OpenAPITypeAgent, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordFromReq))
	// rag
	mid.Sub("openapi").RegWithAPIType(openAPI, "/rag/chat", http.MethodPost, openapi.ChatRag, "文本问答OpenAPI", constant.OpenAPITypeRag, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordFromReq))
	// model
	mid.Sub("openapi").RegWithAPIType(openAPI, "/model/list", http.MethodGet, openapi.ListModels, "模型列表查询OpenAPI", constant.OpenAPITypeModel, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// workflow
	mid.Sub("openapi").RegWithAPIType(openAPI, "/workflow/run", http.MethodPost, openapi.WorkflowRun, "工作流OpenAPI", constant.OpenAPITypeWorkflow, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/workflow/file/upload", http.MethodPost, openapi.WorkflowFileUpload, "工作流OpenAPI文件上传", constant.OpenAPITypeChatflow, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// chatflow
	mid.Sub("openapi").RegWithAPIType(openAPI, "/chatflow/conversation", http.MethodPost, openapi.CreateChatflowConversation, "对话流创建对话OpenAPI", constant.OpenAPITypeChatflow, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/chatflow/conversation", http.MethodDelete, openapi.DeleteChatflowConversation, "对话流删除会话OpenAPI", constant.OpenAPITypeChatflow, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/chatflow/conversation/list", http.MethodPost, openapi.GetChatflowConversationList, "对话流获取会话列表OpenAPI", constant.OpenAPITypeChatflow, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/chatflow/conversation/message/list", http.MethodPost, openapi.GetConversationMessageList, "对话流根据conversationId获取历史对话", constant.OpenAPITypeChatflow, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/chatflow/chat", http.MethodPost, openapi.ChatflowChat, "对话流OpenAPI", constant.OpenAPITypeChatflow, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/chatflow/file/upload", http.MethodPost, openapi.ChatflowFileUpload, "对话流OpenAPI文件上传", constant.OpenAPITypeChatflow, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// knowledge
	mid.Sub("openapi").RegWithAPIType(openAPI, "/file/upload/direct", http.MethodPost, openapi.DirectUploadFiles, "直接上传文件", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge", http.MethodPost, openapi.CreateKnowledge, "新建知识库", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthModelByUuid([]string{"embeddingModelInfo.modelId", "knowledgeGraph.llmModelId"}))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge", http.MethodPut, openapi.UpdateKnowledge, "更新知识库", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge", http.MethodDelete, openapi.DeleteKnowledge, "删除知识库", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/select", http.MethodPost, openapi.GetKnowledgeSelect, "查询知识库列表", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/config", http.MethodGet, openapi.GetDocConfig, "获取文档配置信息", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/list", http.MethodPost, openapi.GetDocList, "获取文档列表", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/import", http.MethodPost, openapi.ImportDoc, "上传文档", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/update/config", http.MethodPost, openapi.UpdateDocConfig, "更新文档配置", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/import/tip", http.MethodGet, openapi.GetDocImportTip, "获取知识库文档上传状态", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/export", http.MethodPost, openapi.ExportKnowledgeDoc, "知识库文档导出", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/export/record/list", http.MethodGet, openapi.GetKnowledgeExportRecordList, "获取知识库导出记录列表", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/export/record", http.MethodDelete, openapi.DeleteKnowledgeExportRecord, "删除知识库库导出记录", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc", http.MethodDelete, openapi.DeleteDoc, "删除文档", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/hit", http.MethodPost, openapi.KnowledgeHit, "知识库命中测试", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthModelByUuid([]string{"knowledgeMatchParams.rerankModelId"}))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/segment/list", http.MethodGet, openapi.GetDocSegmentList, "获取文档切分结果", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeView))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/segment/child/list", http.MethodGet, openapi.GetDocChildSegmentList, "获取子分段列表", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeView))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/segment/status/update", http.MethodGet, openapi.UpdateDocSegmentStatus, "更新分段状态", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/segment/create", http.MethodPost, openapi.CreateDocSegment, "新增文档切片", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/segment/batch/create", http.MethodPost, openapi.BatchCreateDocSegment, "批量新增文档切片", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/segment/update", http.MethodPost, openapi.UpdateDocSegment, "更新文档切片", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/segment/delete", http.MethodDelete, openapi.DeleteDocSegment, "删除文档切片", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledgeDoc("docId", middleware.KnowledgeEdit))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/knowledge/doc/segment/image/upload", http.MethodPost, openapi.UploadDocSegmentImage, "分段图片上传", constant.OpenAPITypeKnowledge, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))

	// mcp server
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/sse", http.MethodGet, openapi.GetMCPServerSSE, "新建MCP服务sse连接", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/message", http.MethodPost, openapi.GetMCPServerMessage, "获取MCP服务sse消息", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/streamable", http.MethodGet, openapi.GetMCPServerStreamable, "获取MCP服务streamable消息(GET)", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/streamable", http.MethodPost, openapi.GetMCPServerStreamable, "获取MCP服务streamable消息(POST)", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))

	// oauth
	mid.Sub("openapi").Reg(openAPI, "/oauth/jwks", http.MethodGet, openapi.OAuthJWKS, "JWT公钥")
	mid.Sub("openapi").Reg(openAPI, "/oauth/login", http.MethodGet, openapi.OAuthLogin, "OAuth登录授权")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/authorize", http.MethodGet, openapi.OAuthAuthorize, "获取授权码")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/token", http.MethodPost, openapi.OAuthToken, "授权码获取token")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/token/refresh", http.MethodPost, openapi.OAuthRefresh, "刷新Access Token")
	mid.Sub("openapi").Reg(openAPI, "/.well-known/openid-configuration", http.MethodGet, openapi.OAuthConfig, "返回Endpoint配置")
	// oauth user
	mid.Sub("openapi").Reg(openAPI, "/oauth/userinfo", http.MethodGet, openapi.OAuthGetUserInfo, "OAuth获取用户信息", middleware.JWTOAuthAccess)

	// wga — 通用智能体
	// 配置与资源
	// mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/config", http.MethodPut, openapi.UpdateWGAConfig, "更新WGA配置", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/resource/select", http.MethodGet, openapi.GetWGAResourceSelect, "获取WGA可选资源", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// 对话管理
	mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/conversation", http.MethodPost, openapi.CreateWGAConversation, "创建WGA对话", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.AuthModelByUuid([]string{"modelUuid"}), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/conversation", http.MethodDelete, openapi.DeleteWGAConversation, "删除WGA对话", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/conversation/list", http.MethodGet, openapi.GetWGAConversationList, "WGA对话列表", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/conversation/detail", http.MethodGet, openapi.GetWGAConversationDetail, "WGA对话详情", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// 聊天交互
	mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/conversation/chat", http.MethodPost, openapi.WGAConversationChat, "WGA对话流", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.AuthModelByUuid([]string{"modelUuid"}), middleware.APIKeyRecord(middleware.RecordStreamType))
	// 人工介入
	// mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/question/reply", http.MethodPost, openapi.WGAReplyQuestion, "WGA回复提问", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/question/reject", http.MethodPost, openapi.WGARejectQuestion, "WGA拒绝提问", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	// 工作区文件
	mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/conversation/workspace", http.MethodGet, openapi.GetWGAWorkspace, "WGA工作区目录树", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
	mid.Sub("openapi").RegWithAPIType(openAPI, "/wga/conversation/workspace/download", http.MethodGet, openapi.WGAWorkspaceDownload, "WGA工作区文件下载", constant.OpenAPITypeWGA, middleware.AuthOpenAPIKey(), middleware.APIKeyRecord(middleware.RecordNonStreamType))
}
