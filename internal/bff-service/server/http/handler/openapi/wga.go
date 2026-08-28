package openapi

import (
	"fmt"
	"strings"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// allowedAgentIds 从 WGA 配置中收集所有合法的子智能体 ID（排除 Skill 相关内部智能体）
func allowedAgentIds() []string {
	var ids []string
	for _, agent := range config.WgaCfg().SubAgents {
		if strings.HasPrefix(agent.AgentID, "Skill") {
			continue // Skill Chat/Import/Preview 等内部智能体不对外开放
		}
		ids = append(ids, agent.AgentID)
	}
	return ids
}

// isAllowedAgentId 检查 agentId 是否合法：空字符串（Supervisor）或配置中定义的合法子智能体
func isAllowedAgentId(agentId string) bool {
	agentId = strings.TrimSpace(agentId)
	if agentId == "" {
		return true // Supervisor 默认路由
	}
	for _, id := range allowedAgentIds() {
		if id == agentId {
			return true
		}
	}
	return false
}

// UpdateWGAConfig
//
//	@Tags			openapi
//	@Summary		更新WGA资源配置
//	@Description	更新通用智能体工具配置
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateGeneralAgentConfigReq	true	"更新WGA配置请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/wga/config [put]
// func UpdateWGAConfig(ctx *gin.Context) {
//	var req request.UpdateGeneralAgentConfigReq
//	if !gin_util.Bind(ctx, &req) {
//		return
//	}
//	err := service.UpdateGeneralAgentConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
//	gin_util.Response(ctx, nil, err)
// }

// GetWGAResourceSelect
//
//	@Tags			openapi
//	@Summary		获取WGA可选资源列表
//	@Description	获取所有可选资源（MCP、工作流、技能、助手、知识库、知识网络）
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.GeneralAgentResourceSelectList}
//	@Router			/wga/resource/select [get]
// func GetWGAResourceSelect(ctx *gin.Context) {
//	resp, err := service.GetGeneralAgentResourceSelect(ctx, getUserID(ctx), getOrgID(ctx), "")
//	gin_util.Response(ctx, resp, err)
// }

// CreateWGAConversation
//
//	@Tags			openapi
//	@Summary		创建WGA对话
//	@Description	创建通用智能体对话（通过 modelUuid 指定模型）
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIWgaCreateConversationReq	true	"创建对话请求参数"
//	@Success		200		{object}	response.Response{data=response.CreateGeneralAgentConversationResp}
//	@Router			/wga/conversation [post]
func CreateWGAConversation(ctx *gin.Context) {
	var req request.OpenAPIWgaCreateConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}

	// 通过 modelUuid 解析模型配置
	modelConfig, err := service.ResolveModelConfigByUuid(ctx, req.ModelUuid)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	convReq := request.CreateGeneralAgentConversationReq{
		Title:       req.Title,
		ModelConfig: modelConfig,
	}
	resp, err := service.CreateGeneralAgentConversation(ctx, getUserID(ctx), getOrgID(ctx), convReq)
	gin_util.Response(ctx, resp, err)
}

// DeleteWGAConversation
//
//	@Tags			openapi
//	@Summary		删除WGA对话
//	@Description	删除通用智能体对话
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteGeneralAgentConversationReq	true	"删除对话请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/wga/conversation [delete]
func DeleteWGAConversation(ctx *gin.Context) {
	var req request.DeleteGeneralAgentConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteGeneralAgentConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// GetWGAConversationList
//
//	@Tags			openapi
//	@Summary		WGA对话列表
//	@Description	获取通用智能体对话历史列表
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		json
//	@Param			pageNo		query		int	true	"页码，从1开始"
//	@Param			pageSize	query		int	true	"每页条数"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.GeneralAgentConversationInfo}}
//	@Router			/wga/conversation/list [get]
func GetWGAConversationList(ctx *gin.Context) {
	var req request.GetGeneralAgentConversationListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetGeneralAgentConversationList(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetWGAConversationDetail
//
//	@Tags			openapi
//	@Summary		WGA对话详情
//	@Description	获取指定会话的对话详情，包含AG-UI事件流历史
//	@Security		OpenAPIKey
//	@Produce		json
//	@Param			threadId	query		string	true	"对话ID"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.GeneralAgentConversationDetailInfo}}
//	@Router			/wga/conversation/detail [get]
func GetWGAConversationDetail(ctx *gin.Context) {
	var req request.GetGeneralAgentConversationDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetGeneralAgentConversationDetail(ctx, getUserID(ctx), getOrgID(ctx), req.ThreadID)
	gin_util.Response(ctx, resp, err)
}

// WGAConversationChat
//
//	@Tags			openapi
//	@Summary		WGA对话流
//	@Description	发送消息并接收AG-UI SSE事件流（modelUuid 必填）
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			data	body		request.OpenAPIWgaConversationChatReq	true	"对话流请求参数"
//	@Success		200		{object}	string									"SSE流式返回"
//	@Router			/wga/conversation/chat [post]
func WGAConversationChat(ctx *gin.Context) {
	var req request.OpenAPIWgaConversationChatReq
	if !gin_util.Bind(ctx, &req) {
		return
	}

	// 校验 agentId 是否合法（空 = Supervisor，或配置中定义的子智能体；Skill 类不开放）
	if !isAllowedAgentId(req.AgentID) {
		gin_util.Response(ctx, nil, fmt.Errorf("agentId '%s' is not allowed", strings.TrimSpace(req.AgentID)))
		return
	}

	// modelUuid 为必填，通过 uuid 解析模型配置并更新对话
	modelConfig, err := service.ResolveModelConfigByUuid(ctx, req.ModelUuid)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if err := service.UpdateGeneralAgentConversationConfig(ctx, getUserID(ctx), getOrgID(ctx),
		request.UpdateGeneralAgentConversationConfigReq{
			ThreadID:    req.ThreadID,
			ModelConfig: modelConfig,
		}); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	// 转换为 service 层请求结构体
	chatReq := request.GeneralAgentConversationChatReq{
		AgentID:  req.AgentID,
		ThreadID: req.ThreadID,
		Messages: req.Messages,
	}
	err = service.GeneralAgentConversationChat(ctx, getUserID(ctx), getOrgID(ctx), "", chatReq, false)
	if err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// WGAReplyQuestion
//
//	@Tags			openapi
//	@Summary		WGA回复提问
//	@Description	回复智能体提问（Human-in-the-Loop），解除AI阻塞等待
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GeneralAgentReplyQuestionReq	true	"回复提问请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/wga/question/reply [post]
// func WGAReplyQuestion(ctx *gin.Context) {
// 	var req request.GeneralAgentReplyQuestionReq
// 	if !gin_util.Bind(ctx, &req) {
// 		return
// 	}
// 	err := service.GeneralAgentReplyQuestion(ctx, req.RunID, req.QuestionID, req.Answers)
// 	gin_util.Response(ctx, nil, err)
// }

// WGARejectQuestion
//
//	@Tags			openapi
//	@Summary		WGA拒绝提问
//	@Description	拒绝智能体提问（Human-in-the-Loop），AI将收到RejectedError
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GeneralAgentRejectQuestionReq	true	"拒绝提问请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/wga/question/reject [post]
// func WGARejectQuestion(ctx *gin.Context) {
// 	var req request.GeneralAgentRejectQuestionReq
// 	if !gin_util.Bind(ctx, &req) {
// 		return
// 	}
// 	err := service.GeneralAgentRejectQuestion(ctx, req.RunID, req.QuestionID)
// 	gin_util.Response(ctx, nil, err)
// }

// GetWGAWorkspace
//
//	@Tags			openapi
//	@Summary		WGA工作区目录树
//	@Description	获取工作区目录结构和文件列表
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		json
//	@Param			threadId	query		string	true	"对话ID"
//	@Param			runId		query		string	false	"运行ID"
//	@Success		200			{object}	response.Response{data=response.GeneralAgentWorkspaceResp}
//	@Router			/wga/conversation/workspace [get]
func GetWGAWorkspace(ctx *gin.Context) {
	var req request.GeneralAgentWorkspaceReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GeneralAgentWorkspaceInfo(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// WGAWorkspaceDownload
//
//	@Tags			openapi
//	@Summary		WGA工作区文件下载
//	@Description	下载工作区中的文件
//	@Security		OpenAPIKey
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			threadId	query	string	true	"对话ID"
//	@Param			runId		query	string	false	"运行ID"
//	@Param			path		query	string	false	"文件路径"
//	@Success		200			{file}	stream
//	@Router			/wga/conversation/workspace/download [get]
func WGAWorkspaceDownload(ctx *gin.Context) {
	var req request.GeneralAgentWorkspaceDownloadReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	fileName, data, err := service.GeneralAgentWorkspaceDownload(ctx, getUserID(ctx), getOrgID(ctx), req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.ResponseAttachment(ctx, fileName, data)
}
