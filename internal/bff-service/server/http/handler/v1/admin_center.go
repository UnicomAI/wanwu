package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// AdminKnowledgePageList
//
//	@Tags			admin_center.knowledge
//	@Summary		管理员中心知识库全局列表
//	@Description	管理员中心知识库全局列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminKnowledgePageListReq	true	"知识库全局列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminKnowledge}
//	@Router			/admin/center/knowledge/page/list [post]
func AdminKnowledgePageList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AdminKnowledgePageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}

	err := service.FillOrgIds(ctx, userId, orgId, &req.AdminUserSelect)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.AdminKnowledgePageList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminKnowledgeBase
//
//	@Tags			admin_center.knowledge
//	@Summary		管理员中心知识库详情
//	@Description	管理员中心知识库详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminKnowledgeDetailReq	true	"知识库详情参数"
//	@Success		200		{object}	response.Response{data=response.AdminKnowledgeBase}
//	@Router			/admin/center/knowledge/base [post]
func AdminKnowledgeBase(ctx *gin.Context) {
	var req request.AdminKnowledgeDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminSelectKnowledgeBase(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminKnowledgeFileList
//
//	@Tags			admin_center.knowledge
//	@Summary		管理员中心知识库文件列表
//	@Description	管理员中心知识库文件列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocListReq	true	"知识库详情参数"
//	@Success		200		{object}	response.PageResult{list=response.ListDocResp}
//	@Router			/admin/center/knowledge/file/list [post]
func AdminKnowledgeFileList(ctx *gin.Context) {
	var req request.DocListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetDocList(ctx, "", "", &req)
	gin_util.Response(ctx, resp, err)
}

// AdminKnowledgeQAPairList
//
//	@Tags			admin_center.knowledge
//	@Summary		管理员中心获取问答对列表
//	@Description	管理员中心获取问答对列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.KnowledgeQAPairListReq	true	"问答对列表查询请求参数"
//	@Success		200		{object}	response.PageResult{list=response.ListKnowledgeQAPairResp}
//	@Router			/admin/center/knowledge/qa/pair/list [post]
func AdminKnowledgeQAPairList(ctx *gin.Context) {
	var req request.KnowledgeQAPairListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeQAPairList(ctx, "", "", &req)
	gin_util.Response(ctx, resp, err)
}

// AdminKnowledgeFileDetail
//
//	@Tags			admin_center.knowledge
//	@Summary		管理员中心知识库文档详情
//	@Description	管理员中心知识库文档详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminKnowledgeFileDetailReq	true	"知识库文档详情参数"
//	@Success		200		{object}	response.Response{data=response.DocSegmentResp}
//	@Router			/admin/center/knowledge/file/detail [post]
func AdminKnowledgeFileDetail(ctx *gin.Context) {
	var req request.AdminKnowledgeFileDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminKnowledgeFileDetail(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminWorkflowPageList
//
//	@Tags			admin_center.workflow
//	@Summary		工作流分页列表
//	@Description	工作流分页列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminWorkflowPageListReq	true	"工作流分页列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminWorkflow}
//	@Router			/admin/center/workflow/page/list [post]
func AdminWorkflowPageList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AdminWorkflowPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.FillOrgIds(ctx, userId, orgId, &req.AdminUserSelect)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.AdminWorkflowPageList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminSkillPageList
//
//	@Tags			admin_center.skill
//	@Summary		skill分页列表
//	@Description	skill分页列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminSkillPageListReq	true	"skill分页列表分页列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminSkillDetail}
//	@Router			/admin/center/skill/page/list [post]
func AdminSkillPageList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AdminSkillPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.FillOrgIds(ctx, userId, orgId, &req.AdminUserSelect)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.AdminSkillPageList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminSkillBase
//
//	@Tags			admin_center.skill
//	@Summary		skill基础信息
//	@Description	skill基础信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminSkillDetailReq	true	"skill基础信息参数"
//	@Success		200		{object}	response.Response{data=response.AdminAppBaseInfo}
//	@Router			/admin/center/skill/base [post]
func AdminSkillBase(ctx *gin.Context) {
	var req request.AdminSkillDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminSkillBase(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminSkillDetail
//
//	@Tags			admin_center.skill
//	@Summary		skill详情
//	@Description	skill详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminSkillDetailReq	true	"skill详情参数"
//	@Success		200		{object}	response.Response{data=response.PublishedSkillDetail}
//	@Router			/admin/center/skill/detail [post]
func AdminSkillDetail(ctx *gin.Context) {
	var req request.AdminSkillDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminSkillDetail(ctx, &req, true)
	gin_util.Response(ctx, resp, err)
}

// AdminSkillVersionList
//
//	@Tags			admin_center.skill
//	@Summary		获取我发布skill版本列表
//	@Description	获取我发布的skill版本历史列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminSkillDetailReq	true	"skill详情参数"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.SkillVersionInfo}}
//	@Router			/admin/center/skill/version/list [get]
func AdminSkillVersionList(ctx *gin.Context) {
	var req request.AdminSkillDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminSkillVersionList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminModelPageList
//
//	@Tags			admin_center.model
//	@Summary		管理员中心模型全局列表
//	@Description	管理员中心模型全局列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminModelPageListReq	true	"模型全局列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminModel}
//	@Router			/admin/center/model/page/list [post]
func AdminModelPageList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AdminModelPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.FillOrgIds(ctx, userId, orgId, &req.AdminUserSelect)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.AdminModelPageList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminModelBase
//
//	@Tags			admin_center.model
//	@Summary		管理员中心模型基础信息
//	@Description	管理员中心模型基础信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminModelDetailReq	true	"模型详情参数"
//	@Success		200		{object}	response.Response{data=response.AdminModelBase}
//	@Router			/admin/center/model/base [post]
func AdminModelBase(ctx *gin.Context) {
	var req request.AdminModelDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminModelBase(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminModelDetail
//
//	@Tags			admin_center.model
//	@Summary		管理员中心模型详情
//	@Description	管理员中心模型详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminModelDetailReq	true	"模型详情参数"
//	@Success		200		{object}	response.Response{data=response.ModelInfo}
//	@Router			/admin/center/model/detail [post]
func AdminModelDetail(ctx *gin.Context) {
	var req request.AdminModelDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetModelById(ctx, &request.GetModelRequest{
		BaseModelRequest: request.BaseModelRequest{ModelId: req.ModelId},
	})
	gin_util.Response(ctx, resp, err)
}

// AdminRagPageList
//
//	@Tags			admin_center.rag
//	@Summary		管理员中心知识问答全局列表
//	@Description	管理员中心知识问答全局列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminRagPageListReq	true	"知识问答全局列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminRag}
//	@Router			/admin/center/rag/page/list [post]
func AdminRagPageList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AdminRagPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.FillOrgIds(ctx, userId, orgId, &req.AdminUserSelect)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.AdminRagPageList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminRagBase
//
//	@Tags			admin_center.rag
//	@Summary		管理员中心知识问答基础信息
//	@Description	管理员中心知识问答基础信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminRagDetailReq	true	"知识问答基本信息参数"
//	@Success		200		{object}	response.Response{data=response.AdminRagBase}
//	@Router			/admin/center/rag/base [post]
func AdminRagBase(ctx *gin.Context) {
	var req request.AdminRagDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminRagBase(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminRagDetail
//
//	@Tags			admin_center.rag
//	@Summary		管理员中心知识问答详情
//	@Description	管理员中心知识问答详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminRagDetailReq	true	"知识问答详情参数"
//	@Success		200		{object}	response.Response{data=response.AdminRagDetail}
//	@Router			/admin/center/rag/detail [post]
func AdminRagDetail(ctx *gin.Context) {
	var req request.AdminRagDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminRagDetail(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminMCPPageList
//
//	@Tags			admin.mcp
//	@Summary		管理员中心MCP全局列表
//	@Description	管理员中心MCP全局列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminMCPPageListReq	true	"MCP全局列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminMCP}
//	@Router			/admin/center/mcp/page/list [post]
func AdminMCPPageList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AdminMCPPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.FillOrgIds(ctx, userId, orgId, &req.AdminUserSelect)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.AdminMCPPageList(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminMCPBase
//
//	@Tags			admin.mcp
//	@Summary		管理员中心MCP基础信息
//	@Description	管理员中心MCP基础信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminMCPBaseReq	true	"MCP基础信息参数"
//	@Success		200		{object}	response.Response{data=response.AdminMCPBase}
//	@Router			/admin/center/mcp/base [post]
func AdminMCPBase(ctx *gin.Context) {
	var req request.AdminMCPBaseReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminMCPBase(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminMCPDetail
//
//	@Tags			admin.mcp
//	@Summary		管理员中心导入MCP详情
//	@Description	管理员中心导入MCP详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminMCPDetailReq	true	"导入MCP详情参数"
//	@Success		200		{object}	response.Response{data=response.MCPDetail}
//	@Router			/admin/center/mcp/custom/detail [post]
func AdminMCPDetail(ctx *gin.Context) {
	var req request.AdminMCPDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminMCPDetail(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminMCPServerDetail
//
//	@Tags			admin.mcp
//	@Summary		管理员中心创建MCP详情
//	@Description	管理员中心创建MCP详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminMCPDetailReq	true	"创建MCP详情参数"
//	@Success		200		{object}	response.Response{data=response.MCPServerDetail}
//	@Router			/admin/center/mcp/server/detail [post]
func AdminMCPServerDetail(ctx *gin.Context) {
	var req request.AdminMCPDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminMCPServerDetail(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminMCPToolList
//
//	@Tags			admin.mcp
//	@Summary		管理员中心获取MCP Tool列表
//	@Description	管理员中心获取MCP Tool列表
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminMCPToolListReq	true	"mcp工具列表请求参数"
//	@Success		200		{object}	response.Response{data=response.MCPToolList}
//	@Router			/admin/center/mcp/tool/list [post]
func AdminMCPToolList(ctx *gin.Context) {
	var req request.AdminMCPToolListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminMCPToolList(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminToolPageList
//
//	@Tags			admin.tool
//	@Summary		管理员中心工具全局列表
//	@Description	管理员中心工具全局列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminToolPageListReq	true	"工具全局列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminTool}
//	@Router			/admin/center/tool/page/list [post]
func AdminToolPageList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AdminToolPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.FillOrgIds(ctx, userId, orgId, &req.AdminUserSelect)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.AdminToolPageList(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminToolBase
//
//	@Tags			admin.tool
//	@Summary		管理员中心工具基础信息
//	@Description	管理员中心工具基础信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminToolBaseReq	true	"工具基础信息参数"
//	@Success		200		{object}	response.Response{data=response.AdminToolBase}
//	@Router			/admin/center/tool/base [post]
func AdminToolBase(ctx *gin.Context) {
	var req request.AdminToolBaseReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminToolBase(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminToolDetail
//
//	@Tags			admin.tool
//	@Summary		管理员中心工具详情
//	@Description	管理员中心工具详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminToolDetailReq	true	"工具详情参数"
//	@Success		200		{object}	response.Response{data=response.CustomToolDetail}
//	@Router			/admin/center/tool/detail [post]
func AdminToolDetail(ctx *gin.Context) {
	var req request.AdminToolDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminToolDetail(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminPromptPageList
//
//	@Tags			admin.prompt
//	@Summary		管理员中心提示词全局列表
//	@Description	管理员中心提示词全局列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminPromptPageListReq	true	"提示词全局列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminPrompt}
//	@Router			/admin/center/prompt/page/list [post]
func AdminPromptPageList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AdminPromptPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.FillOrgIds(ctx, userId, orgId, &req.AdminUserSelect)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp, err := service.AdminPromptPageList(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminPromptBase
//
//	@Tags			admin.prompt
//	@Summary		管理员中心提示词基础信息
//	@Description	管理员中心提示词基础信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminPromptBaseReq	true	"MCP基础信息参数"
//	@Success		200		{object}	response.Response{data=response.AdminPromptBase}
//	@Router			/admin/center/prompt/base [post]
func AdminPromptBase(ctx *gin.Context) {
	var req request.AdminPromptBaseReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminPromptBase(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminPromptDetail
//
//	@Tags			admin.prompt
//	@Summary		管理员中心提示词详情
//	@Description	管理员中心提示词详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminPromptDetailReq	true	"提示词详情参数"
//	@Success		200		{object}	response.Response{data=response.CustomPrompt}
//	@Router			/admin/center/prompt/detail [post]
func AdminPromptDetail(ctx *gin.Context) {
	var req request.AdminPromptDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminPromptDetail(ctx, req)
	gin_util.Response(ctx, resp, err)
}

// AdminAssistantPageList
//
//	@Tags			admin_center.assistant
//	@Summary		管理员中心智能体全局列表
//	@Description	管理员中心智能体全局列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminAssistantPageListReq	true	"智能体全局列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminAssistant}
//	@Router			/admin/center/assistant/page/list [post]
func AdminAssistantPageList(ctx *gin.Context) {
	var req request.AdminAssistantPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.FillOrgIds(ctx, getUserID(ctx), getOrgID(ctx), &req.AdminUserSelect); err != nil {
		gin_util.ResponseErr(ctx, err)
		return
	}
	resp, err := service.AdminAssistantPageList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminAssistantBase
//
//	@Tags			admin_center.assistant
//	@Summary		管理员中心智能体基础信息
//	@Description	管理员中心智能体基础信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminAssistantDetailReq	true	"智能体基础信息参数"
//	@Success		200		{object}	response.Response{data=response.AdminAssistantBase}
//	@Router			/admin/center/assistant/base [post]
func AdminAssistantBase(ctx *gin.Context) {
	var req request.AdminAssistantDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminAssistantBase(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminAssistantDetail
//
//	@Tags			admin_center.assistant
//	@Summary		管理员中心智能体详情
//	@Description	管理员中心智能体详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminAssistantDetailReq	true	"智能体详情参数"
//	@Success		200		{object}	response.Response{data=response.AdminAssistantDetail}
//	@Router			/admin/center/assistant/detail [post]
func AdminAssistantDetail(ctx *gin.Context) {
	var req request.AdminAssistantDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminAssistantDetail(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminSensitiveWordPageList
//
//	@Tags			admin_center.sensitive
//	@Summary		管理员中心敏感词全局列表
//	@Description	管理员中心敏感词全局列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminSensitiveWordPageListReq	true	"敏感词全局列表参数"
//	@Success		200		{object}	response.PageResult{list=response.AdminSensitiveWord}
//	@Router			/admin/center/sensitive/page/list [post]
func AdminSensitiveWordPageList(ctx *gin.Context) {
	var req request.AdminSensitiveWordPageListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.FillOrgIds(ctx, getUserID(ctx), getOrgID(ctx), &req.AdminUserSelect); err != nil {
		gin_util.ResponseErr(ctx, err)
		return
	}
	resp, err := service.AdminSensitiveWordPageList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminSensitiveWordBase
//
//	@Tags			admin_center.sensitive
//	@Summary		管理员中心敏感词基础信息
//	@Description	返回敏感词表的名称、备注、类型、创建时间、更新时间、创建人、所属组织
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminSensitiveWordBaseReq	true	"敏感词基础信息参数"
//	@Success		200		{object}	response.Response{data=response.AdminSensitiveWordBase}
//	@Router			/admin/center/sensitive/base [post]
func AdminSensitiveWordBase(ctx *gin.Context) {
	var req request.AdminSensitiveWordBaseReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminSensitiveWordBase(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// AdminSensitiveWordDetail
//
//	@Tags			admin_center.sensitive
//	@Summary		管理员中心敏感词详情
//	@Description	管理员中心敏感词详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AdminSensitiveWordDetailReq	true	"敏感词详情参数"
//	@Success		200		{object}	response.Response{data=response.AdminSensitiveWordDetailResp}
//	@Router			/admin/center/sensitive/detail [post]
func AdminSensitiveWordDetail(ctx *gin.Context) {
	var req request.AdminSensitiveWordDetailReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AdminSensitiveWordDetail(ctx, &req)
	gin_util.Response(ctx, resp, err)
}
