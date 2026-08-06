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
	mid.Sub("exploration.app").Reg(apiV1, "/exploration/app/list", http.MethodGet, v1.GetExplorationAppList, "获取应用广场应用")
	mid.Sub("exploration.app").Reg(apiV1, "/exploration/app/favorite", http.MethodPost, v1.ChangeExplorationAppFavorite, "更改App收藏状态")

	// rag 相关接口
	mid.Sub("exploration.app").Reg(apiV1, "/appspace/rag", http.MethodGet, v1.GetPublishedRag, "获取已发布rag详情", middleware.AuthAppPublish("ragId", constant.AppTypeRag))
	mid.Sub("exploration.app").Reg(apiV1, "/rag/chat", http.MethodPost, v1.ChatPublishedRag, "已发布rag流式接口", middleware.AppHistoryRecord("ragId", constant.AppTypeRag))

	// agent 相关接口
	mid.Sub("exploration.app").Reg(apiV1, "/assistant", http.MethodGet, v1.GetPublishedAssistantInfo, "查看已发布智能体详情")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation", http.MethodPost, v1.ConversationCreate, "创建智能体对话")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation", http.MethodDelete, v1.ConversationDelete, "删除智能体对话")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation/list", http.MethodGet, v1.GetConversationList, "智能体对话列表")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation/detail", http.MethodGet, v1.GetConversationDetailList, "智能体对话详情历史列表")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/pending/conversation", http.MethodPost, v1.GetAssistantPendingConversion, "获取智能体运行中会话")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/stream/connect", http.MethodPost, v1.AssistantConversionStreamConnect, "智能体流式问答断开后重连")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/stream/cancel", http.MethodPost, v1.AssistantConversionStreamCancel, "智能体流式问答手动停止")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/conversation/clear", http.MethodDelete, v1.ClearPublishedAssistantConversation, "清空已发布智能体对话")
	mid.Sub("exploration.app").Reg(apiV1, "/assistant/stream", http.MethodPost, v1.PublishedAssistantConversionStream, "已发布智能体流式问答", middleware.AppHistoryRecord("assistantId", constant.AppTypeAgent))

	// workflow 相关接口
	mid.Sub("exploration.app").Reg(apiV1, "/workflow/run", http.MethodPost, v1.PublishedWorkflowRun, "已发布工作流运行接口", middleware.AppHistoryRecord("workflow_id", constant.AppTypeWorkflow))
	mid.Sub("exploration.app").Reg(apiV1, "/appspace/workflow/export", http.MethodGet, v1.ExportWorkflow, "导出workflow")
	mid.Sub("exploration.app").Reg(apiV1, "/appspace/workflow/copy", http.MethodPost, v1.CopyWorkflow, "拷贝workflow")

	// chatflow 相关接口
	mid.Sub("exploration.app").Reg(apiV1, "/chatflow/application/list", http.MethodPost, v1.ChatflowApplicationList, "应用广场对话流关联应用", middleware.AppHistoryRecord("workflow_id", constant.AppTypeChatflow))
	mid.Sub("exploration.app").Reg(apiV1, "/chatflow/application/info", http.MethodPost, v1.ChatflowApplicationInfo, "应用广场对话流关联应用信息")
	mid.Sub("exploration.app").Reg(apiV1, "/chatflow/conversation/delete", http.MethodDelete, v1.DeleteChatflowConversation, "删除对话流会话")
	mid.Sub("exploration.app").Reg(apiV1, "/appspace/chatflow/copy", http.MethodPost, v1.CopyChatflow, "拷贝chatflow")

	mid.Sub("exploration.template").Reg(apiV1, "/prompt/template/list", http.MethodGet, v1.GetPromptTemplateList, "获取提示词模板列表")
	mid.Sub("exploration.template").Reg(apiV1, "/prompt/template/detail", http.MethodGet, v1.GetPromptTemplateDetail, "获取提示词模板详情")

	// skill 广场
	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/builtin/list", http.MethodGet, v1.GetSquareBuiltinSkillList, "获取广场内置skill列表")
	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/builtin/detail", http.MethodGet, v1.GetSquareBuiltinSkillDetail, "获取广场内置skill详情")

	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/share/list", http.MethodGet, v1.GetSquareShareSkillList, "共享skill列表")
	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/share", http.MethodPost, v1.ShareSquareSkill, "添加共享skill到资源库")
	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/share/detail", http.MethodGet, v1.GetSquareShareSkillDetail, "获取共享skill详情")
	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/share/download", http.MethodGet, v1.DownloadSquareShareSkill, "下载共享skill")
	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/share/version/list", http.MethodGet, v1.GetSquareShareSkillVersionList, "获取共享skill版本列表")

	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/created/list", http.MethodGet, v1.GetSquareCreatedSkillList, "获取我发布skill列表")
	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/created/detail", http.MethodGet, v1.GetSquareCreatedSkillDetail, "获取我发布skill详情")
	mid.Sub("exploration.skill").Reg(apiV1, "/square/skill/created/version/list", http.MethodGet, v1.GetSquareCreatedSkillVersionList, "获取我发布skill版本列表")

}
