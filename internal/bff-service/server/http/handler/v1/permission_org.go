package v1

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CreateOrg
//
//	@Tags			permission.org
//	@Summary Create subordinate organization
//	@Description Create a subordinate organization of the X-Org-Id organization
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.OrgCreate true "Organization information"
//	@Success		200		{object}	response.Response{data=response.OrgID}
//	@Router			/org [post]
func CreateOrg(ctx *gin.Context) {
	var req request.OrgCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_org_cannot_create"))
		return
	}
	resp, err := service.CreateOrg(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ChangeOrg
//
//	@Tags		permission.org
//	@Summary Edit subordinate organizations
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.OrgUpdate true "Organization information"
//	@Success	200		{object}	response.Response
//	@Router		/org [put]
func ChangeOrg(ctx *gin.Context) {
	var req request.OrgUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_org_cannot_change"))
		return
	}
	err := service.ChangeOrg(ctx, getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteOrg
//
//	@Tags		permission.org
//	@Summary Delete subordinate organizations
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.OrgID true "Organization ID"
//	@Success	200		{object}	response.Response
//	@Router		/org [delete]
func DeleteOrg(ctx *gin.Context) {
	var req request.OrgID
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_org_cannot_delete"))
		return
	}
	err := service.DeleteOrg(ctx, getOrgID(ctx), req.OrgID)
	gin_util.Response(ctx, nil, err)
}

// GetOrgInfo
//
//	@Tags		permission.org
//	@Summary Get organization information
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param orgId query string true "Organization ID"
//	@Success	200		{object}	response.Response{data=response.OrgInfo}
//	@Router		/org/info [get]
func GetOrgInfo(ctx *gin.Context) {
	resp, err := service.GetOrgInfo(ctx, ctx.Query("orgId"))
	gin_util.Response(ctx, resp, err)
}

// GetOrgList
//
//	@Tags			permission.org
//	@Summary Get the list of subordinate organizations
//	@Description Get the list of subordinate organizations of the X-Org-Id organization
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param name query string false "Organization name (fuzzy query)"
//	@Param pageNo query int true "Page number, starting from 1"
//	@Param pageSize query int true "Number of single pages, starting from 1"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.OrgInfo}}
//	@Router			/org/list [get]
func GetOrgList(ctx *gin.Context) {
	resp, err := service.GetOrgList(ctx, getOrgID(ctx), ctx.Query("name"), getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// ChangeOrgStatus
//
//	@Tags		permission.org
//	@Summary Modify the status of subordinate organizations
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.OrgStatus true "Organization information"
//	@Success	200		{object}	response.Response
//	@Router		/org/status [put]
func ChangeOrgStatus(ctx *gin.Context) {
	var req request.OrgStatus
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_org_cannot_change_status"))
		return
	}
	err := service.ChangeOrgStatus(ctx, getOrgID(ctx), req.OrgID.OrgID, req.Status)
	gin_util.Response(ctx, nil, err)
}
