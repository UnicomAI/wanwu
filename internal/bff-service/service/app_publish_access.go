package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CheckAppPublishAccess 校验用户对指定已发布 App 的访问权限。
// 按 PublishType 分支：public 放行；organization 校验同组织；private 校验同用户且同组织。
// app 未发布（查询无记录或 PublishType 为空）视为无权限。
func CheckAppPublishAccess(ctx *gin.Context, appId, appType, userId, orgId string) error {
	resp, err := app.GetAppListByIds(ctx.Request.Context(), &app_service.GetAppListByIdsReq{
		AppIdsList: []string{appId},
		AppType:    appType,
	})
	if err != nil {
		return err
	}

	// 单 id 查询，取第一条非空记录
	var info *app_service.AppInfo
	if resp != nil {
		for _, item := range resp.GetInfos() {
			if item != nil {
				info = item
				break
			}
		}
	}
	if info == nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_app_publish_no_permission")
	}

	// 发布范围权限判断，逻辑对齐 app-service canAccessApp（orm/client.go:201）
	switch info.GetPublishType() {
	case constant.AppPublishPublic:
		return nil
	case constant.AppPublishOrganization:
		if info.GetOrgId() == orgId {
			return nil
		}
	case constant.AppPublishPrivate:
		if info.GetUserId() == userId && info.GetOrgId() == orgId {
			return nil
		}
	default:
		// 未发布（PublishType 为空）或其他未知类型
	}
	return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_app_publish_no_permission")
}
