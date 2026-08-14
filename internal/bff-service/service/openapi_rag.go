package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// GetRagListForOpenAPI 知识问答列表，字段以 uuid 命名并带发布类型与版本号
func GetRagListForOpenAPI(ctx *gin.Context, userID, orgID, name string) (*response.OpenAPIRagListResponse, error) {
	listResp, err := rag.ListRag(ctx.Request.Context(), &rag_service.RagListReq{
		Name: name,
		Identity: &rag_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(listResp.RagInfos) == 0 {
		return &response.OpenAPIRagListResponse{List: []response.OpenAPIRagBriefInfo{}}, nil
	}

	// 查询发布状态与最新版本号
	appIds := make([]string, 0, len(listResp.RagInfos))
	for _, item := range listResp.RagInfos {
		appIds = append(appIds, item.AppId)
	}
	appInfosResp, err := app.GetAppListByIds(ctx.Request.Context(), &app_service.GetAppListByIdsReq{
		AppIdsList: appIds,
		AppType:    constant.AppTypeRag,
	})
	if err != nil {
		return nil, err
	}
	publishAppMap := make(map[string]*app_service.AppInfo, len(appInfosResp.Infos))
	for _, appInfo := range appInfosResp.Infos {
		publishAppMap[appInfo.AppId] = appInfo
	}
	versionMap := getAppVersionBatch(ctx, userID, orgID, publishAppMap)

	briefList := make([]response.OpenAPIRagBriefInfo, 0, len(listResp.RagInfos))
	for _, item := range listResp.RagInfos {
		brief := response.OpenAPIRagBriefInfo{
			UUID:      item.AppId,
			Name:      item.Name,
			Desc:      item.Desc,
			Avatar:    cacheAppAvatar(ctx, item.AvatarPath, constant.AppTypeRag),
			CreatedAt: util.Time2Str(item.CreatedAt),
			UpdatedAt: util.Time2Str(item.UpdatedAt),
		}
		if publishInfo, ok := publishAppMap[item.AppId]; ok {
			brief.PublishType = publishInfo.PublishType
		}
		if version, ok := versionMap[item.AppId]; ok {
			brief.Version = version
		}
		briefList = append(briefList, brief)
	}
	return &response.OpenAPIRagListResponse{List: briefList}, nil
}
