package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// AssistantCreate
//
//	@Tags			agent
//	@Summary		创建智能体
//	@Description	创建智能体，填写基本信息，创建完成为草稿状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantCreateReq	true	"智能体基本信息"
//	@Success		200		{object}	response.Response{data=response.AssistantCreateResp}
//	@Router			/assistant [post]
func AssistantCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantCreateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AssistantCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// AssistantUpdate
//
//	@Tags			agent
//	@Summary		修改智能体基本信息
//	@Description	修改智能体基本信息，名称，头像，简介
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantUpdateReq	true	"智能体基本信息参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant [put]
func AssistantUpdate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AssistantUpdate(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// AssistantConfigUpdate
//
//	@Tags			agent
//	@Summary		修改智能体配置信息
//	@Description	修改智能体配置信息，模型配置，知识库配置等等
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantConfig	true	"智能体配置信息参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/config [put]
func AssistantConfigUpdate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantConfig
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AssistantConfigUpdate(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// GetPublishedAssistantInfo
//
//	@Tags			agent
//	@Summary		查看已发布智能体详情
//	@Description	查看已发布智能体详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.AssistantIdRequest	true	"获取智能体信息的请求参数"
//	@Success		200		{object}	response.Response{data=response.Assistant}
//	@Router			/assistant [get]
func GetPublishedAssistantInfo(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantIdRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAssistantInfo(ctx, userId, orgId, req, true)
	gin_util.Response(ctx, resp, err)
}

// GetDraftAssistantInfo
//
//	@Tags			agent
//	@Summary		查看草稿智能体详情
//	@Description	查看草稿智能体详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.AssistantIdRequest	true	"获取智能体信息的请求参数"
//	@Success		200		{object}	response.Response{data=response.Assistant}
//	@Router			/assistant/draft [get]
func GetDraftAssistantInfo(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantIdRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAssistantInfo(ctx, userId, orgId, req, false)
	gin_util.Response(ctx, resp, err)
}

// AssistantCopy
//
//	@Tags			agent
//	@Summary		复制智能体
//	@Description	复制智能体，创建一个新的智能体，基本信息和配置都和原智能体一致
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantIdRequest	true	"智能体id"
//	@Success		200		{object}	response.Response{data=response.AssistantCreateResp}
//	@Router			/assistant/copy [post]
func AssistantCopy(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantIdRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AssistantCopy(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// AssistantWorkFlowCreate
//
//	@Tags			agent
//	@Summary		添加工作流
//	@Description	为智能体绑定已发布的工作流
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantWorkFlowAddRequest	true	"工作流新增参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool/workflow [post]
func AssistantWorkFlowCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantWorkFlowAddRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantWorkFlowCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantWorkFlowDelete
//
//	@Tags			agent
//	@Summary		删除工作流
//	@Description	为智能体解绑工作流
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantWorkFlowDelRequest	true	"工作流id,智能体id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool/workflow [delete]
func AssistantWorkFlowDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantWorkFlowDelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantWorkFlowDelete(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantWorkFlowEnableSwitch
//
//	@Tags			agent
//	@Summary		启用/停用工作流
//	@Description	修改智能体绑定的工作流的启用状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantWorkFlowToolEnableRequest	true	"工作流id,智能体id,开关"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool/workflow/switch [put]
func AssistantWorkFlowEnableSwitch(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantWorkFlowToolEnableRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantWorkFlowEnableSwitch(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantMCPCreate
//
//	@Tags			agent
//	@Summary		添加mcp工具
//	@Description	为智能体绑定已发布的mcp工具
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantMCPToolAddRequest	true	"mcp工具id、mcp类型、智能体id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool/mcp [post]
func AssistantMCPCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantMCPToolAddRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantMCPCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantMCPDelete
//
//	@Tags			agent
//	@Summary		删除mcp
//	@Description	为智能体解绑mcp
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantMCPToolDelRequest	true	"mcp工具id、mcp类型、智能体id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool/mcp [delete]
func AssistantMCPDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantMCPToolDelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantMCPDelete(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantMCPEnableSwitch
//
//	@Tags			agent
//	@Summary		启用/停用 MCP
//	@Description	修改智能体绑定的MCP的启用状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantMCPToolEnableRequest	true	"mcp工具id、mcp类型、智能体id、enable"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool/mcp/switch [put]
func AssistantMCPEnableSwitch(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantMCPToolEnableRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantMCPEnableSwitch(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantToolCreate
//
//	@Tags			agent
//	@Summary		添加自定义、内建工具
//	@Description	为智能体绑定自定义、内建工具
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantToolAddRequest	true	"自定义、内建工具新增参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool [post]
func AssistantToolCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantToolAddRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantToolCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantToolDelete
//
//	@Tags			agent
//	@Summary		删除自定义、内建工具
//	@Description	为智能体解绑自定义、内建工具
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantToolDelRequest	true	"智能体id与自定义、内建工具id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool [delete]
func AssistantToolDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantToolDelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantToolDelete(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantToolEnableSwitch
//
//	@Tags			agent
//	@Summary		启用/停用自定义、内建工具
//	@Description	修改智能体绑定的自定义、内建工具的启用状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantToolEnableRequest	true	"智能体id与自定义、内建工具id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool/switch [put]
func AssistantToolEnableSwitch(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantToolEnableRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantToolEnableSwitch(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantToolConfig
//
//	@Tags			agent
//	@Summary		配置智能体工具
//	@Description	配置智能体工具，包括自定义工具和内置工具
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantToolConfigRequest	true	"智能体工具配置参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/tool/config [put]
func AssistantToolConfig(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantToolConfigRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantToolConfig(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantSkillCreate
//
//	@Tags			agent
//	@Summary		添加技能
//	@Description	为智能体绑定技能
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantSkillAddRequest	true	"技能新增参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/skill [post]
func AssistantSkillCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantSkillAddRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantSkillCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantSkillDelete
//
//	@Tags			agent
//	@Summary		删除技能
//	@Description	为智能体解绑技能
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantSkillDelRequest	true	"智能体id与技能id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/skill [delete]
func AssistantSkillDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantSkillDelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantSkillDelete(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// AssistantSkillEnableSwitch
//
//	@Tags			agent
//	@Summary		启用/停用技能
//	@Description	修改智能体绑定的技能的启用状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantSkillEnableSwitchRequest	true	"智能体id与技能id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/skill/switch [put]
func AssistantSkillEnableSwitch(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantSkillEnableSwitchRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AssistantSkillEnableSwitch(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// ConversationCreate
//
//	@Tags			agent
//	@Summary		创建智能体对话
//	@Description	创建智能体对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversationCreateRequest	true	"智能体对话创建参数"
//	@Success		200		{object}	response.Response{data=response.ConversationCreateResp}
//	@Router			/assistant/conversation [post]
func ConversationCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationCreateRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}

	resp, err := service.ConversationCreate(ctx, userId, orgId, req, constant.ConversationTypePublished)
	gin_util.Response(ctx, resp, err)
}

// ConversationDelete
//
//	@Tags			agent
//	@Summary		删除智能体对话
//	@Description	删除智能体对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversationIdRequest	true	"智能体对话的id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/conversation [delete]
func ConversationDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationIdRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.ConversationDelete(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// ClearPublishedAssistantConversation
//
//	@Tags			agent
//	@Summary		清空已发布智能体对话
//	@Description	清空已发布智能体对话ES数据，不删除会话ID
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversationIdRequest	true	"智能体对话清空参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/conversation/clear [delete]
func ClearPublishedAssistantConversation(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationIdRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.ClearPublishedConversationES(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// GetConversationList
//
//	@Tags			agent
//	@Summary		智能体对话列表
//	@Description	智能体对话列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			assistantId	query		string	true	"智能体id"
//	@Param			pageNo		query		int		true	"页面编号，从1开始"
//	@Param			pageSize	query		int		true	"单页数量，从1开始"
//	@Param			searchText	query		string	false	"检索词"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.ConversationInfo}}
//	@Router			/assistant/conversation/list [get]
func GetConversationList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationGetListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetConversationList(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// GetConversationDetailList
//
//	@Tags			agent
//	@Summary		智能体对话详情历史列表
//	@Description	智能体对话详情历史列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			conversationId	query		string	true	"智能体对话id"
//	@Param			pageNo			query		int		true	"页面编号，从1开始"
//	@Param			pageSize		query		int		true	"单页数量，从1开始"
//	@Success		200				{object}	response.Response{data=response.PageResult{list=[]response.ConversationDetailInfo}}
//	@Router			/assistant/conversation/detail [get]
func GetConversationDetailList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationGetDetailListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetConversationDetailList(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// DraftAssistantConversionStream
//
//	@Tags			agent
//	@Summary		智能体流式问答
//	@Description	智能体流式问答
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversionStreamRequest	true	"智能体流式问答参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/stream/draft [post]
func DraftAssistantConversionStream(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.ConversionStreamRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}

	if !req.IsCompare {
		conversationIdResp, _ := service.GetDraftConversationIdByAssistantID(ctx, userId, orgId, request.ConversationGetListRequest{
			AssistantId: req.AssistantId,
		})

		if conversationIdResp == nil {
			newConversationId, err := service.ConversationCreate(ctx, userId, orgId, request.ConversationCreateRequest{
				AssistantId: req.AssistantId,
				Prompt:      req.Prompt,
			}, constant.ConversationTypeDraft)
			if err != nil {
				gin_util.Response(ctx, nil, err)
			}
			req.ConversationId = newConversationId.ConversationId
		} else {
			req.ConversationId = conversationIdResp.ConversationId
		}
	}

	//启用链接保持
	req.SseHold = true

	if err := service.AssistantConversionStream(ctx, userId, orgId, clientId, req, false, constant.AppStatisticSourceDraft); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// GetAssistantPendingConversion
//
//	@Tags			agent
//	@Summary		获取智能体运行中会话
//	@Description	获取智能体运行中会话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.PendingConversionRequest	true	"获取智能体运行中会话请求参数"
//	@Success		200		{object}	response.Response{data=response.PendingConversationResp}
//	@Router			/assistant/pending/conversation [post]
func GetAssistantPendingConversion(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.PendingConversionRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	conversation, err := service.GetPendingConversation(ctx, userId, orgId, clientId, req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
	}
	gin_util.Response(ctx, conversation, err)
}

// AssistantConversionStreamConnect
//
//	@Tags			agent
//	@Summary		草稿智能体流式问答断开后重连
//	@Description	草稿智能体流式问答断开后重连
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversionStreamConnectRequest	true	"智能体流式问答参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/stream/connect [post]
func AssistantConversionStreamConnect(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.ConversionStreamConnectRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}

	if err := service.AssistantConversionStreamConnect(ctx, userId, orgId, clientId, req); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// AssistantConversionStreamCancel
//
//	@Tags			agent
//	@Summary		智能体流式问答手动停止
//	@Description	智能体流式问答手动停止
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversionStreamCancelRequest	true	"智能体流式问答手动停止参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/stream/cancel [post]
func AssistantConversionStreamCancel(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.ConversionStreamCancelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}

	if err := service.AssistantConversionStreamCancel(ctx, userId, orgId, clientId, req); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, nil, nil)
}

// DraftAssistantConversationDetailList
//
//	@Tags			agent
//	@Summary		草稿智能体对话详情历史列表
//	@Description	草稿智能体对话详情历史列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			assistantId	query		string	true	"智能体id"
//	@Param			pageNo		query		int		true	"页面编号"
//	@Param			pageSize	query		int		true	"单页数量"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.ConversationDetailInfo}}
//	@Router			/assistant/conversation/draft/detail [get]
func DraftAssistantConversationDetailList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationGetListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}

	// 获取会话id
	conversationIdResp, err := service.GetDraftConversationIdByAssistantID(ctx, userId, orgId, request.ConversationGetListRequest{
		AssistantId: req.AssistantId,
	})
	if err != nil || conversationIdResp == nil {
		gin_util.Response(ctx, response.PageResult{List: []response.ConversationDetailInfo{}}, nil)
		return
	}

	// 获取对话详情列表
	resp, err := service.GetConversationDetailList(ctx, userId, orgId, request.ConversationGetDetailListRequest{
		ConversationId: conversationIdResp.ConversationId,
		PageSize:       req.PageSize,
		PageNo:         req.PageNo,
	})
	gin_util.Response(ctx, resp, err)
}

// DraftAssistantConversationDelete
//
//	@Tags			agent
//	@Summary		删除草稿智能体对话
//	@Description	删除草稿智能体对话，传detailId删除单条，不传删除全部
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversationDeleteRequest	true	"智能体对话删除参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/conversation/draft [delete]
func DraftAssistantConversationDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationDeleteRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.DraftConversationDeleteByAssistantID(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// PublishedAssistantConversionStream
//
//	@Tags			agent
//	@Summary		已发布智能体流式问答
//	@Description	已发布智能体流式问答
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversionStreamRequest	true	"智能体流式问答参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/stream [post]
func PublishedAssistantConversionStream(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.ConversionStreamRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	//启用链接保持
	req.SseHold = true
	if err := service.AssistantConversionStream(ctx, userId, orgId, clientId, req, true, constant.AppStatisticSourceWeb); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// MultiAgentCreate
//
//	@Tags			agent
//	@Summary		添加多智能体配置
//	@Description	为智能体绑定已发布的子智能体
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MultiAgentCreateReq	true	"智能体id、子智能体id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/multi-agent [post]
func MultiAgentCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.MultiAgentCreateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.MultiAgentCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// MultiAgentDelete
//
//	@Tags			agent
//	@Summary		删除多智能体配置
//	@Description	为智能体解绑子智能体
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MultiAgentCreateReq	true	"智能体id与子智能体id"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/multi-agent [delete]
func MultiAgentDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.MultiAgentCreateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.MultiAgentDelete(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// MultiAgentEnableSwitch
//
//	@Tags			agent
//	@Summary		启用/停用多智能体配置
//	@Description	为智能体启用/停用子智能体
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MultiAgentEnableSwitchReq	true	"智能体id、子智能体id、开关状态"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/multi-agent/switch [put]
func MultiAgentEnableSwitch(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.MultiAgentEnableSwitchReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.MultiAgentEnableSwitch(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// MultiAgentConfigUpdate
//
//	@Tags			agent
//	@Summary		修改多智能体配置
//	@Description	为智能体修改子智能体描述
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MultiAgentConfigUpdateReq	true	"智能体id、子智能体id、描述"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/multi-agent/config [put]
func MultiAgentConfigUpdate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.MultiAgentConfigUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.MultiAgentConfigUpdate(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}

// GetAssistantSelect
//
//	@Tags			agent
//	@Summary		添加多智能体配置-下拉列表接口
//	@Description	添加多智能体配置-下拉列表接口
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"assistant名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.ExplorationAppInfo}}
//	@Router			/assistant/select [get]
func GetAssistantSelect(ctx *gin.Context) {
	req := request.GetExplorationAppListRequest{
		Name: ctx.Query("name"),
	}
	resp, err := service.GetAssistantSelect(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// AssistantQuestionRecommend
//
//	@Tags			agent
//	@Summary		智能体问题推荐接口
//	@Description	智能体问题推荐接口
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.QuestionRecommendRequest	true	"智能问题推荐请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/question/recommend [post]
func AssistantQuestionRecommend(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.QuestionRecommendRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !req.Trial {
		if err := service.CheckOpenAPIAccess(ctx, req.AssistantId, constant.AppTypeAgent, userId, orgId); err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
	}
	service.AssistantQuestionRecommend(ctx, userId, orgId, &req)
}
