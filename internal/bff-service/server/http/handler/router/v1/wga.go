package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerWGA(apiV1 *gin.RouterGroup) {

	// 通用智能体配置相关接口
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/sub/list", http.MethodGet, v1.GetGeneralAgentSubList, "通用智能体子智能体列表")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/upload/limit", http.MethodGet, v1.GetGeneralAgentUploadLimit, "通用智能体上传文件格式限制")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/config", http.MethodPut, v1.UpdateGeneralAgentConfig, "修改通用智能体配置")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/config", http.MethodGet, v1.GetGeneralAgentConfig, "通用智能体配置")

	// 通用智能体资源相关接口
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/assistant/select", http.MethodGet, v1.GetGeneralAgentAssistantSelect, "通用智能体下拉列表接口")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/tool/select", http.MethodGet, v1.GetGeneralAgentToolSelect, "通用智能体工具下拉列表接口")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/tool/info", http.MethodGet, v1.GetGeneralAgentToolInfo, "通用智能体工具详情")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/mcp/select", http.MethodGet, v1.GetGeneralAgentMCPSelect, "通用智能体MCP下拉接口列表")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/workflow/select", http.MethodGet, v1.GetGeneralAgentWorkflowSelect, "通用智能体Workflow下拉接口列表")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/select", http.MethodGet, v1.GetGeneralAgentSkillSelect, "通用智能体skill下拉接口列表")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/knowledge/select", http.MethodPost, v1.GetGeneralAgentKnowledgeSelect, "通用智能体知识库下拉接口列表")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/ontology/employee/select", http.MethodGet, v1.GetGeneralAgentOntologyEmployeeSelect, "通用智能体数字员工下拉列表")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/resource/select", http.MethodGet, v1.GetGeneralAgentResourceSelect, "通用智能体资源选择列表")

	// 通用智能体对话相关接口
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation", http.MethodPost, v1.CreateGeneralAgentConversation, "创建通用智能体对话")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation", http.MethodDelete, v1.DeleteGeneralAgentConversation, "删除通用智能体对话")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/list", http.MethodGet, v1.GetGeneralAgentConversationList, "通用智能体对话列表")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/detail", http.MethodGet, v1.GetGeneralAgentConversationDetail, "通用智能体对话详情")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/config", http.MethodGet, v1.GetGeneralAgentConversationConfig, "通用智能体对话配置")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/config", http.MethodPut, v1.UpdateGeneralAgentConversationConfig, "修改通用智能体对话配置")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/config/check", http.MethodPost, v1.CheckGeneralAgentConversationConfig, "通用智能体配置检查")

	// 通用智能体workspace相关接口
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/workspace/download", http.MethodGet, v1.GeneralAgentWorkspaceDownload, "通用智能体workspace下载")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/workspace/preview", http.MethodGet, v1.GeneralAgentWorkspacePreview, "通用智能体workspace预览")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/workspace", http.MethodGet, v1.GeneralAgentWorkspaceInfo, "通用智能体workspace目录树")

	// 通用智能体流式对话相关接口
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/chat", http.MethodPost, v1.GeneralAgentConversationChat, "通用智能体流式问答", middleware.TraceWeb(constant.BizModuleWGA, middleware.WithModuleCreatorFromContext()))
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/pending", http.MethodGet, v1.GeneralAgentConversationPending, "通用智能体运行中会话查询")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/connect", http.MethodPost, v1.GeneralAgentConversationConnect, "通用智能体流式问答断线重连")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/conversation/cancel", http.MethodPost, v1.GeneralAgentConversationCancel, "通用智能体流式问答手动停止")

	// Skill对话相关接口
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/conversation", http.MethodPost, v1.CreateGeneralAgentSkillConversation, "Skill创建对话")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/import/conversation", http.MethodPost, v1.ImportGeneralAgentSkillConversation, "导入Skill专用对话")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/convert/conversation", http.MethodPost, v1.ConvertGeneralAgentSkillConversation, "一键转化为Skill专用对话")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/refresh/conversation", http.MethodPost, v1.RefreshGeneralAgentSkillConversation, "刷新Skill专用对话")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/conversation/chat", http.MethodPost, v1.GeneralAgentSkillConversationChat, "Skill对话流", middleware.TraceWeb(constant.BizModuleResourceSkill, middleware.WithModuleCreatorFromContext()))
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/preview/conversation/detail", http.MethodGet, v1.GetGeneralAgentSkillPreviewConversationDetail, "Skill preview对话详情")

	// Skill对话流式相关接口
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/conversation/pending", http.MethodGet, v1.GeneralAgentSkillConversationPending, "Skill对话运行中会话查询")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/conversation/connect", http.MethodPost, v1.GeneralAgentSkillConversationConnect, "Skill对话流式问答断线重连")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/skill/conversation/cancel", http.MethodPost, v1.GeneralAgentSkillConversationCancel, "Skill对话流式问答手动停止")

	// 通用智能体Human-In-The-Loop相关接口
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/question/reply", http.MethodPost, v1.GeneralAgentReplyQuestion, "回答问题")
	mid.Sub("wga.wanwu_bot").Reg(apiV1, "/general/agent/question/reject", http.MethodPost, v1.GeneralAgentRejectQuestion, "拒绝问题")

}
