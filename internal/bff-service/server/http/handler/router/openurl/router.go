package openurl

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/openurl"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func Register(openUrl *gin.RouterGroup) {
	// --- openurl ---
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix", http.MethodGet, openurl.GetUrlAgentDetail, "获取智能体Url信息")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/llm", http.MethodGet, openurl.GetUrlAgentLLM, "获取智能体模型配置")

	// --- conversation ---
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation", http.MethodPost, openurl.UrlConversationCreate, "创建智能体对话")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation", http.MethodDelete, openurl.UrlConversationDelete, "删除智能体对话")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation/clear", http.MethodDelete, openurl.UrlConversationClear, "清空智能体对话")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation/list", http.MethodGet, openurl.GetUrlConversationList, "获取智能体对话列表")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/conversation/detail", http.MethodGet, openurl.GetUrlConversationDetailList, "智能体对话详情历史列表")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/stream", http.MethodPost, openurl.AssistantUrlConversionStream, "智能体流式问答")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/pending/conversation", http.MethodGet, openurl.GetAssistantPendingConversion, "获取智能体运行中会话")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/stream/connect", http.MethodPost, openurl.AssistantConversionStreamConnect, "智能体流式问答断开后重连")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/stream/cancel", http.MethodPost, openurl.AssistantConversionStreamCancel, "智能体流式问答手动停止")
	mid.Sub("openurl").Reg(openUrl, "/agent/:suffix/recommend", http.MethodPost, openurl.AssistantUrlQuestionRecommend, "智能体推荐问题")

	// --- file upload（匿名访问） ---
	mid.Sub("openurl").Reg(openUrl, "/file/upload", http.MethodPost, openurl.UploadFile, "上传文件")
	mid.Sub("openurl").Reg(openUrl, "/file/merge", http.MethodPost, openurl.MergeFile, "合并文件")
	mid.Sub("openurl").Reg(openUrl, "/file/clean", http.MethodPost, openurl.CleanFile, "清除文件")
}
