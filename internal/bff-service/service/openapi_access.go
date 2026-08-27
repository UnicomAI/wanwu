package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CheckOpenAPIAccess 校验 OpenAPI 调用应用权限
// OpenAPI 仅允许调用自己创建的已发布应用（私有、组织内、公开都可以）
func CheckOpenAPIAccess(ctx *gin.Context, appId, appType, userId, orgId string) error {
	appInfo, err := app.GetAppInfo(ctx.Request.Context(), &app_service.GetAppInfoReq{
		AppId:   appId,
		AppType: appType,
	})
	if err != nil {
		return err
	}
	if appInfo.PublishType == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_app_publish_no_permission")
	}
	if appInfo.UserId != userId || appInfo.OrgId != orgId {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_openapi_app_not_owner")
	}
	return nil
}
