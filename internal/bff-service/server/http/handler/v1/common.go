package v1

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// GetUserPermission
//
//	@Tags		common
//	@Summary	获取用户权限
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.UserPermission}
//	@Router		/user/permission [get]
func GetUserPermission(ctx *gin.Context) {
	resp, err := service.GetUserPermission(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetUserInfo
//
//	@Tags		common
//	@Summary	获取用户信息
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		orgId	query		string	true	"组织ID"
//	@Success	200		{object}	response.Response{data=response.UserInfo}
//	@Router		/user/info [get]
func GetUserInfo(ctx *gin.Context) {
	resp, err := service.GetUserInfo(ctx, getUserID(ctx), ctx.Query("orgId"))
	gin_util.Response(ctx, resp, err)
}

// GetOrgSelect
//
//	@Tags		common
//	@Summary	获取用户组织列表（用于下拉选择）
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.Select}
//	@Router		/org/select [get]
func GetOrgSelect(ctx *gin.Context) {
	resp, err := service.GetOrgSelect(ctx, getUserID(ctx))
	gin_util.Response(ctx, resp, err)
}

// UploadAvatar
//
//	@Tags		common
//	@Summary	上传自定义图标
//	@Security	JWT
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		avatar	formData	file	true	"自定义图标（JPG/JPEG/PNG）"
//	@Success	200		{object}	response.Response{data=request.Avatar}
//	@Router		/avatar [post]
func UploadAvatar(ctx *gin.Context) {
	avatarFile, err := ctx.FormFile("avatar")
	if err != nil {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFGeneral, "bff_avatar_upload_error", err.Error())
		return
	}
	avatarObjectPath, err := service.UploadAvatar(ctx, avatarFile)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	avatar := service.CacheAvatar(avatarObjectPath)
	gin_util.Response(ctx, avatar, nil)
}

// SearchDocCenter
//
//	@Tags		common
//	@Summary	查找文档中心内容
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		content	query		string	true	"搜索关键字"
//	@Success	200		{object}	response.Response{data=[]response.DocSearchResp}
//	@Router		/doc_center/search [get]
func SearchDocCenter(ctx *gin.Context) {
	resp, err := service.SearchDocCenter(ctx, ctx.Query("content"))
	gin_util.Response(ctx, resp, err)
}

// GetDocCenterMenu
//
//	@Tags		common
//	@Summary	获取文档中心目录
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=[]response.DocMenu}
//	@Router		/doc_center/menu [get]
func GetDocCenterMenu(ctx *gin.Context) {
	gin_util.Response(ctx, service.GetDocCenterMenu(ctx), nil)
}

// GetDocCenterMarkdown
//
//	@Tags		common
//	@Summary	获取文档中心Markdown文件内容
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		path	query		string	true	"目录path"
//	@Success	200		{object}	response.Response{data=string}
//	@Router		/doc_center/markdown [get]
func GetDocCenterMarkdown(ctx *gin.Context) {
	resp, err := service.GetDocCenterMarkdown(ctx, ctx.Query("path"))
	gin_util.Response(ctx, resp, err)
}

// UpdateUserAvatar
//
//	@Tags			common
//	@Summary		编辑用户头像
//	@Description	更新用户的头像信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UserAvatarUpdate	true	"用户头像信息"
//	@Success		200		{object}	response.Response
//	@Router			/user/avatar [put]
func UpdateUserAvatar(ctx *gin.Context) {
	var req request.UserAvatarUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateUserAvatar(ctx, getUserID(ctx), req.Avatar.Key)
	gin_util.Response(ctx, nil, err)
}

// LoginEmailCheck
//
//	@Tags			common
//	@Summary		非首次登录邮箱校验与绑定
//	@Description	二阶段用户非首次登录邮箱校验与绑定
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Language	header		string					false	"语言"
//	@Param			data		body		request.LoginEmailCheck	true	"登录邮箱信息"
//	@Success		200			{object}	response.Response{data=response.Login}
//	@Router			/user/login [post]
func LoginEmailCheck(ctx *gin.Context) {
	var req request.LoginEmailCheck
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// 二阶段用户非首次登录邮箱校验与绑定
	resp, err := service.LoginEmailCheck(ctx, &req, getLanguage(ctx), getUserID(ctx))
	gin_util.Response(ctx, resp, err)
}

// ChangeUserPasswordByEmail
//
//	@Tags			common
//	@Summary		用户首次登录邮箱校验绑定与修改密码
//	@Description	二阶段用户首次登录邮箱校验绑定与修改密码
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Language	header		string								false	"语言"
//	@Param			data		body		request.ChangeUserPasswordByEmail	true	"密码和邮箱信息"
//	@Success		200			{object}	response.Response{data=response.Login}
//	@Router			/user/login [put]
func ChangeUserPasswordByEmail(ctx *gin.Context) {
	var req request.ChangeUserPasswordByEmail
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.ChangeUserPasswordByEmail(ctx, &req, getLanguage(ctx), getUserID(ctx))
	gin_util.Response(ctx, resp, err)
}

// LoginSendEmailCode
//
//	@Tags		common
//	@Summary	登录邮箱验证码发送
//	@Accept		json
//	@Produce	application/json
//	@Param		data	body		request.LoginSendEmailCode	true	"邮箱地址"
//	@Success	200		{object}	response.Response
//	@Router		/user/login/email/code [post]
func LoginSendEmailCode(ctx *gin.Context) {
	var req request.LoginSendEmailCode
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.LoginSendEmailCode(ctx, req.Email)
	gin_util.Response(ctx, nil, err)
}

// --- internal ---

// 获取当前用户ID
func getUserID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.USER_ID)
}

// 获取当前组织ID
func getOrgID(ctx *gin.Context) string {
	return ctx.GetHeader(gin_util.X_ORG_ID)
}

// 获取客户端ID
func getClientID(ctx *gin.Context) string {
	return ctx.GetHeader(gin_util.X_CLIENT_ID)
}

// 获取当前系统语言
func getLanguage(ctx *gin.Context) string {
	return ctx.GetHeader(gin_util.X_LANGUAGE)
}

// 当前用户是否是当前组织内置管理员角色
func isAdmin(ctx *gin.Context) bool {
	return ctx.GetBool(gin_util.IS_ADMIN)
}

// 当前组织是否是内置顶级【系统】组织
func isSystem(ctx *gin.Context) bool {
	return ctx.GetBool(gin_util.IS_SYSTEM)
}

func getPageNo(ctx *gin.Context) int32 {
	return util.MustI32(ctx.Query(gin_util.PageNo))
}

func getPageSize(ctx *gin.Context) int32 {
	return util.MustI32(ctx.Query(gin_util.PageSize))
}
