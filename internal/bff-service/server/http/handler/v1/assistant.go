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
//	@Summary Create an agent
//	@Description Create an agent, fill in the basic information, and the creation will be in draft state
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AppBriefConfig true "Basic information of the agent"
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
//	@Summary Modify the basic information of the agent
//	@Description Modify the basic information of the agent, name, avatar, and introduction
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantBrief true "Basic information parameters of the agent"
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
//	@Summary Modify agent configuration information
//	@Description Modify agent configuration information, model configuration, knowledge base configuration, etc.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantConfig true "Agent configuration information parameters"
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
//	@Summary View the details of the agent after release
//	@Description View the details of the agent after release
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param assistantId query string true "agent id"
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
//	@Summary View draft agent details
//	@Description View draft agent details
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param assistantId query string true "agent id"
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
//	@Summary Copy the agent
//	@Description Copy the agent and create a new agent. The basic information and configuration are the same as the original agent.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantIdRequest true "Agent ID"
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
//	@Summary Add workflow
//	@Description binds the published workflow to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantWorkFlowAddRequest true "Add parameters to workflow"
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
//	@Summary Delete workflow
//	@Description unbinds the workflow for the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantWorkFlowDelRequest true "workflow id, agent id"
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
//	@Summary Enable/disable workflow
//	@Description Modifies the enabled status of the workflow bound to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantWorkFlowToolEnableRequest true "workflow id, agent id, switch"
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
//	@Summary Add mcp tool
//	@Description Binds the published mcp tool to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantMCPToolAddRequest true "mcp tool id, mcp type, agent id"
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
//	@Summary delete mcp
//	@Description Unbind mcp for the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantMCPToolDelRequest true "mcp tool id, mcp type, agent id"
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
//	@Summary Enable/disable MCP
//	@Description Modifies the enabled status of the MCP bound to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantMCPToolEnableRequest true "mcp tool id, mcp type, agent id, enable"
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
//	@Summary Add custom, built-in tools
//	@Description Bind custom, built-in tools to the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantToolAddRequest true "Customized, built-in tool added parameters"
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
//	@Summary Delete custom and built-in tools
//	@Description Unbundles custom and built-in tools for the agent
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantToolDelRequest true "Agent id and custom and built-in tool id"
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
//	@Summary Enable/disable custom and built-in tools
//	@Description Modifies the customization of intelligent agent binding and the enabled status of built-in tools
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantToolEnableRequest true "Agent id and custom and built-in tool id"
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
//	@Summary Configure the agent tool
//	@Description Configure agent tools, including custom tools and built-in tools
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.AssistantToolConfigRequest true "Agent tool configuration parameters"
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
//	@Summary Create agent dialogue
//	@Description Create an agent dialogue
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.ConversationCreateRequest true "Agent conversation creation parameters"
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
//	@Summary Delete agent dialogue
//	@Description Delete agent dialogue
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.ConversationIdRequest true "ID of the agent conversation"
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
//	@Summary Agent dialogue list
//	@Description Agent dialogue list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param assistantId query string true "agent id"
//	@Param pageNo query int true "Page number, starting from 1"
//	@Param pageSize query int true "Number of single pages, starting from 1"
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
//	@Summary Agent conversation details history list
//	@Description Agent conversation details history list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param conversationId query string true "agent conversation id"
//	@Param pageNo query int true "Page number, starting from 1"
//	@Param pageSize query int true "Number of single pages, starting from 1"
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
//	@Summary Agent Streaming Q&A
//	@Description Agent Streaming Q&A
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.ConversionStreamRequest true "Agent streaming question and answer parameters"
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
