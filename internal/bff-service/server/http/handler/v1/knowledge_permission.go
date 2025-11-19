package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// SelectKnowledgeOrg
//
//	@Tags			knowledge.permission
//	@Summary		知识库组织列表 [EN] @Summary Knowledge base organization list
//	@Description	知识库组织列表 [EN] @Description Knowledge base organization list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.KnowledgeOrgSelectReq	true	"知识库组织列表请求参数" [EN] @Param data query request.KnowledgeOrgSelectReq true "Knowledge base organization list request parameters"
//	@Success		200		{object}	response.Response{data=response.KnowOrgInfo}
//	@Router			/knowledge/org [get]
func SelectKnowledgeOrg(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeOrgSelectReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.SelectKnowledgeOrg(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// SelectKnowledgeUserPermit
//
//	@Tags			knowledge.permission
//	@Summary		知识库用户权限列表 [EN] @Summary Knowledge base user permission list
//	@Description	知识库用户权限列表 [EN] @Description Knowledge base user permission list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.KnowledgeUserSelectReq	true	"知识库用户权限列表请求参数" [EN] @Param data query request.KnowledgeUserSelectReq true "Knowledge base user permission list request parameters"
//	@Success		200		{object}	response.Response{data=response.KnowledgeUserPermissionResp}
//	@Router			/knowledge/user [get]
func SelectKnowledgeUserPermit(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeUserSelectReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.SelectKnowledgePermissionUser(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// SelectKnowledgeUserNoPermit
//
//	@Tags			knowledge.permission
//	@Summary		没有知识库用户列表 [EN] @Summary There is no list of knowledge base users
//	@Description	没有知识库用户列表 [EN] @Description There is no knowledge base user list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.KnowledgeUserNoPermitSelectReq	true	"没有知识库用户列表请求参数" [EN] @Param data query request.KnowledgeUserNoPermitSelectReq true "There is no knowledge base user list request parameter"
//	@Success		200		{object}	response.Response{data=response.KnowOrgUserInfoResp}
//	@Router			/knowledge/user/no/permit [get]
func SelectKnowledgeUserNoPermit(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeUserNoPermitSelectReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.SelectKnowledgeNoPermissionUser(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// AddKnowledgeUser
//
//	@Tags			knowledge.permission
//	@Summary		增加知识库用户 [EN] @Summary Add knowledge base users
//	@Description	增加知识库用户 [EN] @Description Add knowledge base users
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeUserAddReq	true	"增加知识库用户请求参数" [EN] @Param data body request.KnowledgeUserAddReq true "Add knowledge base user request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/user/add [post]
func AddKnowledgeUser(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeUserAddReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AddKnowledgeUser(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// EditKnowledgeUser
//
//	@Tags			knowledge.permission
//	@Summary		修改知识库用户 [EN] @Summary Modify knowledge base user
//	@Description	修改知识库用户 [EN] @Description Modify knowledge base user
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeUserEditReq	true	"修改知识库用户请求参数" [EN] @Param data body request.KnowledgeUserEditReq true "Modify knowledge base user request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/user/edit [post]
func EditKnowledgeUser(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeUserEditReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.EditKnowledgeUser(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteKnowledgeUser
//
//	@Tags			knowledge.permission
//	@Summary		删除知识库用户 [EN] @Summary Delete knowledge base user
//	@Description	删除知识库用户 [EN] @Description Delete knowledge base user
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeUserDeleteReq	true	"删除知识库用户请求参数" [EN] @Param data body request.KnowledgeUserDeleteReq true "Delete knowledge base user request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/user/delete [delete]
func DeleteKnowledgeUser(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeUserDeleteReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteKnowledgeUser(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// TransferKnowledgeUserAdmin
//
//	@Tags			knowledge.permission
//	@Summary		转让管理员权限 [EN] @Summary Transfer administrator rights
//	@Description	转让管理员权限 [EN] @Description Transfer administrator rights
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeTransferUserAdminReq	true	"转让管理员权限请求参数" [EN] @Param data body request.KnowledgeTransferUserAdminReq true "Transfer administrator permissions request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/user/admin/transfer [post]
func TransferKnowledgeUserAdmin(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeTransferUserAdminReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.TransferKnowledgeAdminUser(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
