package v1

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// GetRoleTemplate
//
//	@Tags			admin_center
//	@Summary		获取角色模板（用于创建角色）
//	@Description	获取当前用户在X-Org-Id组织的角色模板
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			orgId	query		string	true	"组织ID"
//	@Success		200		{object}	response.Response{data=response.RoleTemplate}
//	@Router			/role/template [get]
func GetRoleTemplate(ctx *gin.Context) {
	resp, err := service.GetRoleTemplateExcludeAdminCenter(ctx, getUserID(ctx), ctx.Query("orgId"))
	gin_util.Response(ctx, resp, err)
}

// CreateRole
//
//	@Tags			admin_center
//	@Summary		创建角色
//	@Description	创建角色（全局角色或组织角色）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RoleCreate	true	"角色信息"
//	@Success		200		{object}	response.Response{data=response.RoleID}
//	@Router			/role [post]
func CreateRole(ctx *gin.Context) {
	var req request.RoleCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// 全局角色只能由系统管理员创建
	if req.IsGlobal && req.OrgID != config.TopOrgID {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_global_role_cannot_operate"))
		return
	}
	resp, err := service.CreateRole(ctx, getUserID(ctx), getOrgID(ctx), req.OrgID, &req)
	gin_util.Response(ctx, resp, err)
}

// ChangeRole
//
//	@Tags			admin_center
//	@Summary		编辑角色
//	@Description	编辑角色（区分全局和组织角色）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RoleUpdate	true	"角色信息"
//	@Success		200		{object}	response.Response
//	@Router			/role [put]
func ChangeRole(ctx *gin.Context) {
	var req request.RoleUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// 全局角色只能由系统管理员编辑
	if req.IsGlobal && req.OrgID != config.TopOrgID {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_global_role_cannot_operate"))
		return
	}
	err := service.ChangeRole(ctx, getUserID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteRole
//
//	@Tags			admin_center
//	@Summary		删除角色
//	@Description	删除角色（roleId全局唯一，无需传orgId）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RoleDelete	true	"角色ID"
//	@Success		200		{object}	response.Response
//	@Router			/role [delete]
func DeleteRole(ctx *gin.Context) {
	var req request.RoleDelete
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// 全局角色只能由系统管理员删除
	if !checkGlobalRolePermission(ctx, req.RoleID, req.OrgID) {
		return
	}
	err := service.DeleteRole(ctx, &req)
	gin_util.Response(ctx, nil, err)
}

// GetRoleInfo
//
//	@Tags			admin_center
//	@Summary		获取角色信息
//	@Description	获取指定角色信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			roleId	query		string	true	"角色ID"
//	@Param			orgId	query		string	true	"组织ID"
//	@Success		200		{object}	response.Response{data=response.RoleInfo}
//	@Router			/role/info [get]
func GetRoleInfo(ctx *gin.Context) {
	resp, err := service.GetRoleInfo(ctx, getUserID(ctx), ctx.Query("orgId"), ctx.Query("roleId"))
	gin_util.Response(ctx, resp, err)
}

// GetRoleList
//
//	@Tags			admin_center
//	@Summary		获取角色列表
//	@Description	获取指定组织的角色列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"角色名(模糊查询)"
//	@Param			orgId	query		string	true	"组织ID"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.RoleInfo}}
//	@Router			/role/list [get]
func GetRoleList(ctx *gin.Context) {
	resp, err := service.GetRoleList(ctx, getUserID(ctx), ctx.Query("orgId"), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// ChangeRoleStatus
//
//	@Tags			admin_center
//	@Summary		修改角色状态
//	@Description	修改指定组织的指定角色的角色状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RoleStatus	true	"角色信息"
//	@Success		200		{object}	response.Response
//	@Router			/role/status [put]
func ChangeRoleStatus(ctx *gin.Context) {
	var req request.RoleStatus
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// 全局角色只能由系统管理员启停
	if !checkGlobalRolePermission(ctx, req.RoleID, req.OrgID) {
		return
	}
	err := service.ChangeRoleStatus(ctx, req.OrgID, req.RoleID, req.Status)
	gin_util.Response(ctx, nil, err)
}

// GetRoleUsers
//
//	@Tags			admin_center
//	@Summary		获取角色关联用户列表
//	@Description	获取指定角色关联的用户信息（用户名、电话、所在组织）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			roleId		query		string	true	"角色ID"
//	@Param			name		query		string	false	"用户名模糊搜索"
//	@Param			orgId		query		string	true	"组织ID"
//	@Param			pageNo		query		int		true	"页面编号，从1开始"
//	@Param			pageSize	query		int		true	"单页数量，从1开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.RoleUser}}
//	@Router			/role/users [get]
func GetRoleUsers(ctx *gin.Context) {
	resp, err := service.GetRoleUsers(ctx, ctx.Query("roleId"), ctx.Query("name"), ctx.Query("orgId"), getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// RemoveRoleUser
//
//	@Tags			admin_center
//	@Summary		移除角色关联用户
//	@Description	移除指定用户的指定角色关联
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RoleUserRemove	true	"角色用户信息"
//	@Success		200		{object}	response.Response
//	@Router			/role/user [delete]
func RemoveRoleUser(ctx *gin.Context) {
	var req request.RoleUserRemove
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// 全局角色只能由系统管理员移除用户
	if !checkGlobalRolePermission(ctx, req.RoleID, req.OrgID) {
		return
	}
	err := service.RemoveRoleUser(ctx, req.RoleID, req.UserID, req.OrgID)
	gin_util.Response(ctx, nil, err)
}

// checkGlobalRolePermission 检查全局角色的操作权限：
// 如果目标角色是全局角色，则只有系统管理员（isSystem && isAdmin）才能操作
// 返回 true 表示有权限，false 表示无权限（已写入错误响应）
func checkGlobalRolePermission(ctx *gin.Context, roleID string, orgID string) bool {
	isGlobal, err := service.IsGlobalRole(ctx, orgID, roleID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return false
	}
	if isGlobal && orgID != config.TopOrgID {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_global_role_cannot_operate"))
		return false
	}
	return true
}
