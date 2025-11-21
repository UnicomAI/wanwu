package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerExploration(apiV1 *gin.RouterGroup) {
	mid.Sub("exploration").Reg(apiV1, "/exploration/app/list", http.MethodGet, v1.GetExplorationAppList, "获取应用广场应用")
	mid.Sub("exploration").Reg(apiV1, "/exploration/app/favorite", http.MethodPost, v1.ChangeExplorationAppFavorite, "更改App收藏状态")

	// rag related interfaces
	mid.Sub("exploration").Reg(apiV1, "/appspace/rag", http.MethodGet, v1.GetRag, "获取rag详情")
	mid.Sub("exploration").Reg(apiV1, "/rag/chat", http.MethodPost, v1.ChatRag, "rag流式接口", middleware.AppHistoryRecord("ragId", constant.AppTypeRag))
	// agent related interface
	mid.Sub("exploration").Reg(apiV1, "/assistant", http.MethodGet, v1.GetAssistantInfo, "查看智能体详情")
	mid.Sub("exploration").Reg(apiV1, "/assistant/conversation", http.MethodPost, v1.ConversationCreate, "创建智能体对话")
	mid.Sub("exploration").Reg(apiV1, "/assistant/conversation", http.MethodDelete, v1.ConversationDelete, "删除智能体对话")
	mid.Sub("exploration").Reg(apiV1, "/assistant/conversation/list", http.MethodGet, v1.GetConversationList, "智能体对话列表")
	mid.Sub("exploration").Reg(apiV1, "/assistant/conversation/detail", http.MethodGet, v1.GetConversationDetailList, "智能体对话详情历史列表")
	mid.Sub("exploration").Reg(apiV1, "/assistant/stream", http.MethodPost, v1.AssistantConversionStream, "智能体流式问答", middleware.AppHistoryRecord("assistantId", constant.AppTypeAgent))
}
