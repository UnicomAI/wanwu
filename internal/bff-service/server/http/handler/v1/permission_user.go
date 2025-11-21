package v1

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CreateUser
//
//	@Tags			permission.user
//	@Summary Create user
//	@Description Create a user and join the X-Org-Id organization at the same time; create a user from the system perspective, do not join any organization, and cannot assign roles.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UserCreate true "User information"
//	@Success		200		{object}	response.Response{data=response.UserID}
//	@Router			/user [post]
func CreateUser(ctx *gin.Context) {
	var req request.UserCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_add_other"))
		return
	}
	resp, err := service.CreateUser(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ChangeUser
//
//	@Tags			permission.user
//	@Summary Edit user
//	@Description Edit the user of the X-Org-Id organization; editing the user from the system perspective cannot assign roles
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UserUpdate true "User information"
//	@Success		200		{object}	response.Response
//	@Router			/user [put]
func ChangeUser(ctx *gin.Context) {
	var req request.UserUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_change_other"))
		return
	}
	err := service.ChangeUser(ctx, getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteUser
//
//	@Tags			permission.user
//	@Summary Delete user
//	@Description Remove the user from the X-Org-Id organization; delete the user from the system perspective
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UserID true "User ID"
//	@Success		200		{object}	response.Response
//	@Router			/user [delete]
func DeleteUser(ctx *gin.Context) {
	var req request.UserID
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// delete
	if isSystem(ctx) {
		if !isAdmin(ctx) {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_delete"))
			return
		}
		err := service.DeleteUser(ctx, req.UserID)
		gin_util.Response(ctx, nil, err)
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_delete_other"))
		return
	}
	// remove from org
	err := service.RemoveOrgUser(ctx, getOrgID(ctx), req.UserID)
	gin_util.Response(ctx, nil, err)
}

// GetUserList
//
//	@Tags			permission.user
//	@Summary Get the user list
//	@Description Get the user list of the X-Org-Id organization; get the list of all users in the system from the system perspective
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param name query string false "Username (fuzzy query)"
//	@Param pageNo query int true "Page number, starting from 1"
//	@Param pageSize query int true "Number of single pages, starting from 1"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.UserInfo}}
//	@Router			/user/list [get]
func GetUserList(ctx *gin.Context) {
	resp, err := service.GetUserList(ctx, getOrgID(ctx), ctx.Query("name"), getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// ChangeUserStatus
//
//	@Tags		permission.user
//	@Summary Modify user status
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.UserStatus true "User information"
//	@Success	200		{object}	response.Response
//	@Router		/user/status [put]
func ChangeUserStatus(ctx *gin.Context) {
	var req request.UserStatus
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_change_status"))
		return
	}
	err := service.ChangeUserStatus(ctx, req.UserID.UserID, getOrgID(ctx), req.Status)
	gin_util.Response(ctx, nil, err)
}

// ChangeUserPassword
//
//	@Tags		permission.user
//	@Summary Modify user password (by individual)
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.UserPassword true "User information"
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
	err := service.ChangeUserPassword(ctx, req.UserID.UserID, req.OldPassword, req.NewPassword)
	gin_util.Response(ctx, nil, err)
}

// AdminChangeUserPassword
//
//	@Tags		permission.user
//	@Summary Reset user password (by administrator)
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.UserPasswordByAdmin true "User information"
//	@Success	200		{object}	response.Response
//	@Router		/user/admin/password [put]
func AdminChangeUserPassword(ctx *gin.Context) {
	var req request.UserPasswordByAdmin
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_change_other_password"))
		return
	}
	err := service.AdminChangeUserPassword(ctx, req.UserID.UserID, req.Password)
	gin_util.Response(ctx, nil, err)
}

// GetOrgUserNotSelect
//
//	@Tags			permission.user
//	@Summary Get the list of users who are not in the organization (for drop-down selection)
//	@Description Get the user list of non-X-Org-Id organizations
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param name query string false "Username (fuzzy query)"
//	@Success		200		{object}	response.Response{data=response.Select}
//	@Router			/org/other/select [get]
func GetOrgUserNotSelect(ctx *gin.Context) {
	resp, err := service.GetOrgUserNotSelect(ctx, getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetRoleSelect
//
//	@Tags			permission.user
//	@Summary Get the list of organizational roles (for drop-down selection)
//	@Description Get the role list of the X-Org-Id organization
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.Select}
//	@Router			/role/select [get]
func GetRoleSelect(ctx *gin.Context) {
	resp, err := service.GetRoleSelect(ctx, getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// AddOrgUser
//
//	@Tags			permission.user
//	@Summary Invite users to join the organization
//	@Description Add users of X-Org-Id organization
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.OrgUserAdd true "User-role"
//	@Success		200		{object}	response.Response
//	@Router			/org/user [post]
func AddOrgUser(ctx *gin.Context) {
	var req request.OrgUserAdd
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AddOrgUser(ctx, getOrgID(ctx), req.UserID.UserID, req.RoleID)
	gin_util.Response(ctx, nil, err)
}
