package v1

import (
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CreateUser
//
//	@Tags			admin_center
//	@Summary		创建用户
//	@Description	创建用户，同时加入X-Org-Id组织；在系统视角下创建用户，不加入任何组织，也不能分配角色
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UserCreate	true	"用户信息"
//	@Success		200		{object}	response.Response{data=response.UserID}
//	@Router			/user [post]
func CreateUser(ctx *gin.Context) {
	var req request.UserCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateUser(ctx, getUserID(ctx), req.OrgID, &req)
	gin_util.Response(ctx, resp, err)
}

// ChangeUser
//
//	@Tags			admin_center
//	@Summary		编辑用户
//	@Description	编辑X-Org-Id组织的用户；在系统视角下编辑用户，不能分配角色
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UserUpdate	true	"用户信息"
//	@Success		200		{object}	response.Response
//	@Router			/user [put]
func ChangeUser(ctx *gin.Context) {
	var req request.UserUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ChangeUser(ctx, req.OrgID, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteUser
//
//	@Tags			admin_center
//	@Summary		删除用户
//	@Description	从X-Org-Id组织将用户移除；在系统视角下为删除用户
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UserDelete	true	"用户ID"
//	@Success		200		{object}	response.Response
//	@Router			/user [delete]
func DeleteUser(ctx *gin.Context) {
	var req request.UserDelete
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// delete
	if req.OrgID == config.TopOrgID {
		if !isAdmin(ctx) {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_delete"))
			return
		}
		err := service.DeleteUser(ctx, req.UserID)
		gin_util.Response(ctx, nil, err)
		return
	}
	// 组织视角：从指定组织移除用户，校验对目标组织的管理权
	if !service.IsAdminInOrgs(ctx, getUserID(ctx), req.OrgID) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_delete_other"))
		return
	}
	// remove from org
	err := service.RemoveOrgUser(ctx, req.OrgID, req.UserID)
	gin_util.Response(ctx, nil, err)
}

// GetUserList
//
//	@Tags			admin_center
//	@Summary		获取用户列表
//	@Description	获取X-Org-Id组织的用户列表；在系统视角下获取系统内全部用户列表；name同时模糊匹配用户名和邮箱
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"用户名或邮箱(模糊查询)"
//	@Param			orgId		query		string	true	"组织ID"
//	@Param			roleIds		query		string	false	"角色ID(逗号分隔)"
//	@Param			pageNo		query		int		true	"页面编号，从1开始"
//	@Param			pageSize	query		int		true	"单页数量，从1开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.UserInfo}}
//	@Router			/user/list [get]
func GetUserList(ctx *gin.Context) {
	roleIdsStr := ctx.Query("roleIds")
	var roleIDs []string
	if roleIdsStr != "" {
		roleIDs = strings.Split(roleIdsStr, ",")
	}
	name := ctx.Query("name")
	orgId := ctx.Query("orgId")
	resp, err := service.GetUserList(ctx, orgId, name, roleIDs, getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// ChangeUserStatus
//
//	@Tags		admin_center
//	@Summary	修改用户状态
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.UserStatus	true	"用户信息"
//	@Success	200		{object}	response.Response
//	@Router		/user/status [put]
func ChangeUserStatus(ctx *gin.Context) {
	var req request.UserStatus
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ChangeUserStatus(ctx, req.UserID, req.OrgID, req.Status)
	gin_util.Response(ctx, nil, err)
}

// ChangeUserPassword
//
//	@Tags		admin_center
//	@Summary	修改用户密码（by 个人）
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.UserPassword	true	"用户信息"
//	@Success	200		{object}	response.Response
//	@Router		/user/password [put]
func ChangeUserPassword(ctx *gin.Context) {
	var req request.UserPassword
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if req.UserID.UserID != getUserID(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_change_other_password"))
		return
	}
	err := service.ChangeUserPassword(ctx, req.UserID.UserID, &req)
	gin_util.Response(ctx, nil, err)
}

// AdminChangeUserPassword
//
//	@Tags		admin_center
//	@Summary	重置用户密码（by 管理员）
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.UserPasswordByAdmin	true	"用户信息"
//	@Success	200		{object}	response.Response
//	@Router		/user/admin/password [put]
func AdminChangeUserPassword(ctx *gin.Context) {
	var req request.UserPasswordByAdmin
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AdminChangeUserPassword(ctx, req.UserID, &req)
	gin_util.Response(ctx, nil, err)
}

// GetOrgUserNotSelect
//
//	@Tags			admin_center
//	@Summary		获取不在组织中用户列表（用于下拉选择）
//	@Description	获取非X-Org-Id组织的用户列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"用户名(模糊查询)"
//	@Param			orgId	query		string	true	"组织ID"
//	@Success		200		{object}	response.Response{data=response.Select}
//	@Router			/org/other/select [get]
func GetOrgUserNotSelect(ctx *gin.Context) {
	resp, err := service.GetOrgUserNotSelect(ctx, ctx.Query("orgId"), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetRoleSelect
//
//	@Tags			admin_center
//	@Summary		获取组织角色列表（用于下拉选择）
//	@Description	获取指定组织的角色列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			orgId	query		string	true	"组织ID"
//	@Success		200		{object}	response.Response{data=response.Select}
//	@Router			/role/select [get]
func GetRoleSelect(ctx *gin.Context) {
	resp, err := service.GetRoleSelect(ctx, ctx.Query("orgId"))
	gin_util.Response(ctx, resp, err)
}

// AddOrgUser
//
//	@Tags			admin_center
//	@Summary		邀请用户加入组织
//	@Description	增加指定组织的用户
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OrgUserAdd	true	"用户-角色"
//	@Success		200		{object}	response.Response
//	@Router			/org/user [post]
func AddOrgUser(ctx *gin.Context) {
	var req request.OrgUserAdd
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AddOrgUser(ctx, req.OrgID, req.UserID, req.RoleID)
	gin_util.Response(ctx, nil, err)
}

// CreateUserByFile
//
//	@Tags			admin_center
//	@Summary		批量导入用户
//	@Description	通过Excel文件批量导入用户
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			orgId	query		string	true	"组织ID"
//	@Param			file	formData	file	true	"用户Excel文件"
//	@Success		200		{object}	response.Response{data=response.UserBatchImportResult}
//	@Router			/user/batch [post]
func CreateUserByFile(ctx *gin.Context) {
	orgID := ctx.PostForm("orgId")
	resp, err := service.CreateUserByFile(ctx, getUserID(ctx), orgID)
	gin_util.Response(ctx, resp, err)
}
