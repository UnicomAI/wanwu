package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetAssistantTemplateList
//
//	@Tags			agent
//	@Summary		智能体模板列表 [EN] @Summary Agent template list
//	@Description	智能体模板列表 [EN] @Description Agent template list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			category	query		string	false	"智能体类型"	Enums(gov,industry,edu,medical) [EN] @Param category query string false "agent type" Enums(gov,industry,edu,medical)
//	@Param			name		query		string	false	"智能体名称" [EN] @Param name query string false "agent name"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.AssistantTemplateInfo}}
//	@Router			/assistant/template/list [get]
func GetAssistantTemplateList(ctx *gin.Context) {
	resp, err := service.GetAssistantTemplateList(ctx, ctx.Query("category"), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)

}

// GetAssistantTemplate
//
//	@Tags			agent
//	@Summary		获取智能体模板 [EN] @Summary Get the agent template
//	@Description	获取智能体模板 [EN] @Description Get the agent template
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			assistantTemplateId	query		string	true	"智能体模板ID" [EN] @Param assistantTemplateId query string true "Agent template ID"
//	@Success		200					{object}	response.Response{data=response.AssistantTemplateInfo}
//	@Router			/assistant/template [get]
func GetAssistantTemplate(ctx *gin.Context) {
	resp, err := service.GetAssistantTemplate(ctx, ctx.Query("assistantTemplateId"))
	gin_util.Response(ctx, resp, err)
}

// AssistantTemplateCreate
//
//	@Tags			agent
//	@Summary		复制创建智能体 [EN] @Summary Copy and create the agent
//	@Description	通过复制创建智能体 [EN] @Description Create agent by copying
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AssistantTemplateRequest	true	"复制创建智能体所需信息" [EN] @Param data body request.AssistantTemplateRequest true "Copy the information required to create the agent"
//	@Success		200		{object}	response.Response{data=response.AssistantCreateResp}
//	@Router			/assistant/template [post]
func AssistantTemplateCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AssistantTemplateRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AssistantTemplateCreate(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}
