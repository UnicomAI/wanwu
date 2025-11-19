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
//	@Summary Get user permissions
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
//	@Summary Get user information
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.UserInfo}
//	@Router		/user/info [get]
func GetUserInfo(ctx *gin.Context) {
	resp, err := service.GetUserInfo(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetOrgSelect
//
//	@Tags		common
//	@Summary Get the user organization list (for drop-down selection)
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
//	@Summary Upload custom icon
//	@Security	JWT
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param avatar formData file true "Custom icon (JPG/JPEG/PNG)"
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
	avatar := service.CacheAvatar(ctx, avatarObjectPath, false)
	gin_util.Response(ctx, avatar, nil)
}

// SearchDocCenter
//
//	@Tags		common
//	@Summary Find Document Center content
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param content query string true "Search keyword"
//	@Success	200		{object}	response.Response{data=[]response.DocSearchResp}
//	@Router		/doc_center/search [get]
func SearchDocCenter(ctx *gin.Context) {
	resp, err := service.SearchDocCenter(ctx, ctx.Query("content"))
	gin_util.Response(ctx, resp, err)
}

// GetDocCenterMenu
//
//	@Tags		common
//	@Summary Get the document center directory
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
//	@Summary Get the content of the Markdown file in the Document Center
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param path query string true "directory path"
//	@Success	200		{object}	response.Response{data=string}
//	@Router		/doc_center/markdown [get]
func GetDocCenterMarkdown(ctx *gin.Context) {
	resp, err := service.GetDocCenterMarkdown(ctx, ctx.Query("path"))
	gin_util.Response(ctx, resp, err)
}

// UpdateUserAvatar
//
//	@Tags			common
//	@Summary Edit user avatar
//	@Description updates the user’s avatar information
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UserAvatarUpdate true "User avatar information"
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
//	@Summary Non-first login email verification and binding
//	@Description Second-stage user email verification and binding when not logging in for the first time
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param X-Language header string false "Language"
//	@Param data body request.LoginEmailCheck true "Login email information"
//	@Success		200			{object}	response.Response{data=response.Login}
//	@Router			/user/login [post]
func LoginEmailCheck(ctx *gin.Context) {
	var req request.LoginEmailCheck
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// Second-stage user email verification and binding when not logging in for the first time
	resp, err := service.LoginEmailCheck(ctx, &req, getLanguage(ctx), getUserID(ctx))
	gin_util.Response(ctx, resp, err)
}

// ChangeUserPasswordByEmail
//
//	@Tags			common
//	@Summary User first logs in to email to verify binding and change password
//	@Description In the second stage, users log in to their email address for the first time to verify binding and change passwords.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param X-Language header string false "Language"
//	@Param data body request.ChangeUserPasswordByEmail true "Password and email information"
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
//	@Summary Login email verification code sent
//	@Accept		json
//	@Produce	application/json
//	@Param data body request.LoginSendEmailCode true "Email address"
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

// Get current user ID
func getUserID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.USER_ID)
}

// Get the current organization ID
func getOrgID(ctx *gin.Context) string {
	return ctx.GetHeader(gin_util.X_ORG_ID)
}

// Get client ID
func getClientID(ctx *gin.Context) string {
	return ctx.GetHeader(gin_util.X_CLIENT_ID)
}

// Get the current system language
func getLanguage(ctx *gin.Context) string {
	return ctx.GetHeader(gin_util.X_LANGUAGE)
}

// Whether the current user has a built-in administrator role in the current organization
func isAdmin(ctx *gin.Context) bool {
	return ctx.GetBool(gin_util.IS_ADMIN)
}

// Whether the current organization is a built-in top-level [system] organization
func isSystem(ctx *gin.Context) bool {
	return ctx.GetBool(gin_util.IS_SYSTEM)
}

func getPageNo(ctx *gin.Context) int32 {
	return util.MustI32(ctx.Query(gin_util.PageNo))
}

func getPageSize(ctx *gin.Context) int32 {
	return util.MustI32(ctx.Query(gin_util.PageSize))
}
