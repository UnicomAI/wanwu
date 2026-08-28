package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// GetAgentListForOpenAPI 获取当前用户的智能体列表（OpenAPI 专用）
// 与内部 GetAppSpaceAppList 的区别：返回字段以 UUID 代替 AppId，供外部 API 调用方使用
func GetAgentListForOpenAPI(ctx *gin.Context, userID, orgID, name string) (*response.OpenAPIAgentListResponse, error) {
	// 1. 拉取智能体列表（包含 uuid）
	listResp, err := assistant.GetAssistantListMyAll(ctx.Request.Context(), &assistant_service.GetAssistantListMyAllReq{
		Name: name,
		Identity: &assistant_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(listResp.AssistantInfos) == 0 {
		return &response.OpenAPIAgentListResponse{List: []response.OpenAPIAgentBriefInfo{}}, nil
	}

	// 2. 收集 appId 列表，查询发布状态
	appIds := make([]string, 0, len(listResp.AssistantInfos))
	for _, item := range listResp.AssistantInfos {
		if item.Info != nil {
			appIds = append(appIds, item.Info.AppId)
		}
	}

	appInfosResp, err := app.GetAppListByIds(ctx.Request.Context(), &app_service.GetAppListByIdsReq{
		AppIdsList: appIds,
		AppType:    constant.AppTypeAgent,
	})
	if err != nil {
		return nil, err
	}

	publishAppMap := make(map[string]*app_service.AppInfo, len(appInfosResp.Infos))
	for _, appInfo := range appInfosResp.Infos {
		publishAppMap[appInfo.AppId] = appInfo
	}

	// 3. 批量查询最新版本号
	versionMap := getAppVersionBatch(ctx, userID, orgID, publishAppMap)

	// 4. 组装响应
	briefList := make([]response.OpenAPIAgentBriefInfo, 0, len(listResp.AssistantInfos))
	for _, item := range listResp.AssistantInfos {
		if item.Info == nil {
			continue
		}
		info := item.Info
		brief := response.OpenAPIAgentBriefInfo{
			UUID:      item.Uuid,
			Name:      info.Name,
			Desc:      info.Desc,
			Avatar:    cacheAppAvatar(ctx, info.AvatarPath, constant.AppTypeAgent),
			Category:  item.Category,
			CreatedAt: util.Time2Str(info.CreatedAt),
			UpdatedAt: util.Time2Str(info.UpdatedAt),
		}
		if publishInfo, ok := publishAppMap[info.AppId]; ok {
			brief.PublishType = publishInfo.PublishType
		}
		if version, ok := versionMap[info.AppId]; ok {
			brief.Version = version
		}
		briefList = append(briefList, brief)
	}

	return &response.OpenAPIAgentListResponse{
		List: briefList,
	}, nil
}
