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
//	@Success		200		{object}	response.PageResult{list=response.AdminWorkflowDetail}
//	@Router			/admin/center/workflow/page/list [post]
func AdminWorkflowPageList(ctx *gin.Context) {
	var req request.AdminWorkflowPageListReq
	if !gin_util.Bind(ctx, &req) {
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
	resp, err := service.AdminSkillDetail(ctx, &req)
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
