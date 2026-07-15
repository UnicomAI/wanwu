package service

import (
	"fmt"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
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
		return fmt.Errorf("permission denied: app not published")
	}
	if appInfo.UserId != userId || appInfo.OrgId != orgId {
		return fmt.Errorf("permission denied: openapi can only access app created by yourself")
	}
	return nil
}
