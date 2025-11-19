package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// AssistantCreate
//
//	@Tags			agent
//	@Summary		创建智能体 [EN] @Summary Create an agent
//	@Description	创建智能体，填写基本信息，创建完成为草稿状态 [EN] @Description Create an agent, fill in the basic information, and the creation will be in draft state
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppBriefConfig	true	"智能体基本信息" [EN] @Param data body request.AppBriefConfig true "Basic information of the agent"
//	@Success		200		{object}	response.Response{data=response.AssistantCreateResp}
//	@Router			/assistant [post]
func AssistantCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AppBriefConfig
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AssistantCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// AssistantUpdate
//
//	@Tags			agent
//	@Summary		修改智能体基本信息 [EN] @Summary Modify the basic information of the agent
//	@Description	修改智能体基本信息，名称，头像，简介 [EN] @Description Modify the basic information of the agent, name, avatar, and introduction
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantBrief	true	"智能体基本信息参数" [EN] @Param data body request.AssistantBrief true "Basic information parameters of the agent"
//	@Success		200		{object}	response.Response
//	@Router			/assistant [put]
func AssistantUpdate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantBrief
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AssistantUpdate(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// AssistantConfigUpdate
//
//	@Tags			agent
//	@Summary		修改智能体配置信息 [EN] @Summary Modify agent configuration information
//	@Description	修改智能体配置信息，模型配置，知识库配置等等 [EN] @Description Modify agent configuration information, model configuration, knowledge base configuration, etc.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantConfig	true	"智能体配置信息参数" [EN] @Param data body request.AssistantConfig true "Agent configuration information parameters"
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

// GetAssistantInfo
//
//	@Tags			agent
//	@Summary		查看发布后智能体详情 [EN] @Summary View the details of the agent after release
//	@Description	查看发布后智能体详情 [EN] @Description View the details of the agent after release
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			assistantId	query		string	true	"智能体id" [EN] @Param assistantId query string true "agent id"
//	@Success		200			{object}	response.Response{data=response.Assistant}
//	@Router			/assistant [get]
func GetAssistantInfo(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantIdRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAssistantInfo(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// GetAssistantDraftInfo
//
//	@Tags			agent
//	@Summary		查看草稿智能体详情 [EN] @Summary View draft agent details
//	@Description	查看草稿智能体详情 [EN] @Description View draft agent details
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			assistantId	query		string	true	"智能体id" [EN] @Param assistantId query string true "agent id"
//	@Success		200			{object}	response.Response{data=response.Assistant}
//	@Router			/assistant/draft [get]
func GetAssistantDraftInfo(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantIdRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetAssistantDraftInfo(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// AssistantCopy
//
//	@Tags			agent
//	@Summary		复制智能体 [EN] @Summary Copy the agent
//	@Description	复制智能体，创建一个新的智能体，基本信息和配置都和原智能体一致 [EN] @Description Copy the agent and create a new agent. The basic information and configuration are the same as the original agent.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantIdRequest	true	"智能体id" [EN] @Param data body request.AssistantIdRequest true "Agent ID"
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
//	@Summary		添加工作流 [EN] @Summary Add workflow
//	@Description	为智能体绑定已发布的工作流 [EN] @Description binds the published workflow to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantWorkFlowAddRequest	true	"工作流新增参数" [EN] @Param data body request.AssistantWorkFlowAddRequest true "Add parameters to workflow"
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
//	@Summary		删除工作流 [EN] @Summary Delete workflow
//	@Description	为智能体解绑工作流 [EN] @Description unbinds the workflow for the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantWorkFlowDelRequest	true	"工作流id,智能体id" [EN] @Param data body request.AssistantWorkFlowDelRequest true "workflow id, agent id"
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
//	@Summary		启用/停用工作流 [EN] @Summary Enable/disable workflow
//	@Description	修改智能体绑定的工作流的启用状态 [EN] @Description Modifies the enabled status of the workflow bound to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantWorkFlowToolEnableRequest	true	"工作流id,智能体id,开关" [EN] @Param data body request.AssistantWorkFlowToolEnableRequest true "workflow id, agent id, switch"
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
//	@Summary		添加mcp工具 [EN] @Summary Add mcp tool
//	@Description	为智能体绑定已发布的mcp工具 [EN] @Description Binds the published mcp tool to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantMCPToolAddRequest	true	"mcp工具id、mcp类型、智能体id" [EN] @Param data body request.AssistantMCPToolAddRequest true "mcp tool id, mcp type, agent id"
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
//	@Summary		删除mcp [EN] @Summary delete mcp
//	@Description	为智能体解绑mcp [EN] @Description Unbind mcp for the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantMCPToolDelRequest	true	"mcp工具id、mcp类型、智能体id" [EN] @Param data body request.AssistantMCPToolDelRequest true "mcp tool id, mcp type, agent id"
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
//	@Summary		启用/停用 MCP [EN] @Summary Enable/disable MCP
//	@Description	修改智能体绑定的MCP的启用状态 [EN] @Description Modifies the enabled status of the MCP bound to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantMCPToolEnableRequest	true	"mcp工具id、mcp类型、智能体id、enable" [EN] @Param data body request.AssistantMCPToolEnableRequest true "mcp tool id, mcp type, agent id, enable"
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
//	@Summary		添加自定义、内建工具 [EN] @Summary Add custom, built-in tools
//	@Description	为智能体绑定自定义、内建工具 [EN] @Description Bind custom, built-in tools to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantToolAddRequest	true	"自定义、内建工具新增参数" [EN] @Param data body request.AssistantToolAddRequest true "Customized, built-in tool added parameters"
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
//	@Summary		删除自定义、内建工具 [EN] @Summary Delete custom and built-in tools
//	@Description	为智能体解绑自定义、内建工具 [EN] @Description Unbundles custom and built-in tools for the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantToolDelRequest	true	"智能体id与自定义、内建工具id" [EN] @Param data body request.AssistantToolDelRequest true "Agent id and custom and built-in tool id"
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
//	@Summary		启用/停用自定义、内建工具 [EN] @Summary Enable/disable custom and built-in tools
//	@Description	修改智能体绑定的自定义、内建工具的启用状态 [EN] @Description Modifies the customization of intelligent agent binding and the enabled status of built-in tools
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantToolEnableRequest	true	"智能体id与自定义、内建工具id" [EN] @Param data body request.AssistantToolEnableRequest true "Agent id and custom and built-in tool id"
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
//	@Summary		配置智能体工具 [EN] @Summary Configure the agent tool
//	@Description	配置智能体工具，包括自定义工具和内置工具 [EN] @Description Configure agent tools, including custom tools and built-in tools
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantToolConfigRequest	true	"智能体工具配置参数" [EN] @Param data body request.AssistantToolConfigRequest true "Agent tool configuration parameters"
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

// ConversationCreate
//
//	@Tags			agent
//	@Summary		创建智能体对话 [EN] @Summary Create agent dialogue
//	@Description	创建智能体对话 [EN] @Description Create an agent dialogue
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversationCreateRequest	true	"智能体对话创建参数" [EN] @Param data body request.ConversationCreateRequest true "Agent conversation creation parameters"
//	@Success		200		{object}	response.Response{data=response.ConversationCreateResp}
//	@Router			/assistant/conversation [post]
func ConversationCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationCreateRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.ConversationCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// ConversationDelete
//
//	@Tags			agent
//	@Summary		删除智能体对话 [EN] @Summary Delete agent dialogue
//	@Description	删除智能体对话 [EN] @Description Delete agent dialogue
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversationIdRequest	true	"智能体对话的id" [EN] @Param data body request.ConversationIdRequest true "ID of the agent conversation"
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

// GetConversationList
//
//	@Tags			agent
//	@Summary		智能体对话列表 [EN] @Summary Agent dialogue list
//	@Description	智能体对话列表 [EN] @Description Agent dialogue list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			assistantId	query		string	true	"智能体id" [EN] @Param assistantId query string true "agent id"
//	@Param			pageNo		query		int		true	"页面编号，从1开始" [EN] @Param pageNo query int true "Page number, starting from 1"
//	@Param			pageSize	query		int		true	"单页数量，从1开始" [EN] @Param pageSize query int true "Number of single pages, starting from 1"
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
//	@Summary		智能体对话详情历史列表 [EN] @Summary Agent conversation details history list
//	@Description	智能体对话详情历史列表 [EN] @Description Agent conversation details history list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			conversationId	query		string	true	"智能体对话id" [EN] @Param conversationId query string true "agent conversation id"
//	@Param			pageNo			query		int		true	"页面编号，从1开始" [EN] @Param pageNo query int true "Page number, starting from 1"
//	@Param			pageSize		query		int		true	"单页数量，从1开始" [EN] @Param pageSize query int true "Number of single pages, starting from 1"
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

// AssistantConversionStream
//
//	@Tags			agent
//	@Summary		智能体流式问答 [EN] @Summary Agent Streaming Q&A
//	@Description	智能体流式问答 [EN] @Description Agent Streaming Q&A
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversionStreamRequest	true	"智能体流式问答参数" [EN] @Param data body request.ConversionStreamRequest true "Agent streaming question and answer parameters"
//	@Success		200		{object}	response.Response
//	@Router			/assistant/stream [post]
func AssistantConversionStream(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversionStreamRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.AssistantConversionStream(ctx, userId, orgId, req); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}
