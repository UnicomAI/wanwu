package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateOrg
//
//	@Tags			admin_center
//	@Summary		创建下级组织
//	@Description	创建指定组织的下级组织
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OrgCreate	true	"组织信息"
//	@Success		200		{object}	response.Response{data=response.OrgID}
//	@Router			/org [post]
func CreateOrg(ctx *gin.Context) {
	var req request.OrgCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateOrg(ctx, getUserID(ctx), req.OrgID.OrgID, &req)
	gin_util.Response(ctx, resp, err)
}

// ChangeOrg
//
//	@Tags		admin_center
//	@Summary	编辑下级组织
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.OrgUpdate	true	"组织信息"
//	@Success	200		{object}	response.Response
//	@Router		/org [put]
func ChangeOrg(ctx *gin.Context) {
	var req request.OrgUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ChangeOrg(ctx, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteOrg
//
//	@Tags		admin_center
//	@Summary	删除下级组织
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.OrgID	true	"组织ID"
//	@Success	200		{object}	response.Response
//	@Router		/org [delete]
func DeleteOrg(ctx *gin.Context) {
	var req request.OrgID
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteOrg(ctx, getOrgID(ctx), req.OrgID)
	gin_util.Response(ctx, nil, err)
}

// GetOrgInfo
//
//	@Tags		admin_center
//	@Summary	获取组织信息
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		orgId	query		string	true	"组织ID"
//	@Success	200		{object}	response.Response{data=response.OrgInfo}
//	@Router		/org/info [get]
func GetOrgInfo(ctx *gin.Context) {
	resp, err := service.GetOrgInfo(ctx, ctx.Query("orgId"))
	gin_util.Response(ctx, resp, err)
}

// GetOrgList
//
//	@Tags			admin_center
//	@Summary		获取下级组织列表
//	@Description	获取X-Org-Id组织的下级组织列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"组织名(模糊查询)"
//	@Param			orgId		query		string	true	"组织ID"
//	@Param			pageNo		query		int		true	"页面编号，从1开始"
//	@Param			pageSize	query		int		true	"单页数量，从1开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.OrgInfo}}
//	@Router			/org/list [get]
func GetOrgList(ctx *gin.Context) {
	resp, err := service.GetOrgList(ctx, ctx.Query("orgId"), ctx.Query("name"), getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// ChangeOrgStatus
//
//	@Tags		admin_center
//	@Summary	修改下级组织状态
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.OrgStatus	true	"组织信息"
//	@Success	200		{object}	response.Response
//	@Router		/org/status [put]
func ChangeOrgStatus(ctx *gin.Context) {
	var req request.OrgStatus
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ChangeOrgStatus(ctx, getOrgID(ctx), req.OrgID.OrgID, req.Status)
	gin_util.Response(ctx, nil, err)
}

// GetAdminOrgSubTree
//
//	@Tags			admin_center
//	@Summary		获取管理员组织及下级组织列表
//	@Description	返回当前用户具有组织管理员角色的组织及其下级所有组织的树形结构
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.AdminOrgTreeNode}
//	@Router			/org/tree [get]
func GetAdminOrgSubTree(ctx *gin.Context) {
	resp, err := service.GetAdminOrgSubTree(ctx, getUserID(ctx))
	gin_util.Response(ctx, resp, err)
}
