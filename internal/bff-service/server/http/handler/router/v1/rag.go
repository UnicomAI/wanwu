package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerRag(apiV1 *gin.RouterGroup) {
	mid.Sub("app.rag").Reg(apiV1, "/appspace/rag/list", http.MethodGet, v1.GetRagAppSpaceAppList, "获取文本问答应用列表")
	mid.Sub("app.rag").Reg(apiV1, "/appspace/rag", http.MethodPost, v1.CreateRag, "创建rag")
	mid.Sub("app.rag").Reg(apiV1, "/appspace/rag", http.MethodPut, v1.UpdateRag, "修改rag基本信息")
	mid.Sub("app.rag").Reg(apiV1, "/appspace/rag/config", http.MethodPut, v1.UpdateRagConfig, "修改rag配置信息", middleware.AuthModelByModelId([]string{"modelConfig.modelId", "rerankConfig.modelId", "qaRerankConfig.modelId"}))
	mid.Sub("app.rag").Reg(apiV1, "/appspace/rag", http.MethodDelete, v1.DeleteRag, "删除rag")
	mid.Sub("app.rag").Reg(apiV1, "/appspace/rag/draft", http.MethodGet, v1.GetDraftRag, "获取草稿rag详情")
	mid.Sub("app.rag").Reg(apiV1, "/appspace/rag", http.MethodGet, v1.GetPublishedRag, "获取已发布rag详情", middleware.AuthAppPublish("ragId", constant.AppTypeRag))
	mid.Sub("app.rag").Reg(apiV1, "/appspace/rag/copy", http.MethodPost, v1.CopyRag, "复制rag")
	mid.Sub("app.rag").Reg(apiV1, "/rag/chat/draft", http.MethodPost, v1.ChatDraftRag, "草稿rag流式接口", middleware.TraceWeb(constant.BizModuleAppRag, middleware.WithDraftAppResource(constant.AppTypeRag, "ragId")))
	mid.Sub("app.rag").Reg(apiV1, "/rag/pending/conversation", http.MethodPost, v1.GetRagPendingConversation, "获取知识问答运行中会话")
	mid.Sub("app.rag").Reg(apiV1, "/rag/stream/connect", http.MethodPost, v1.ChatRagStreamConnect, "知识问答流式问答断开后重连", middleware.AuthRagConversationAccess("ragId", "conversationId"))
	mid.Sub("app.rag").Reg(apiV1, "/rag/stream/cancel", http.MethodPost, v1.ChatRagStreamCancel, "知识问答流式问答手动停止")
	mid.Sub("app.rag").Reg(apiV1, "/rag/conversation/draft/detail", http.MethodGet, v1.RagConversationDraftDetailList, "草稿知识问答对话详情历史列表")
	mid.Sub("app.rag").Reg(apiV1, "/rag/conversation/draft", http.MethodDelete, v1.RagConversationDraftDelete, "草稿知识问答对话删除")
	mid.Sub("app.rag").Reg(apiV1, "/rag/upload", http.MethodPost, v1.RagUpload, "文档上传直接传到rag")
}
