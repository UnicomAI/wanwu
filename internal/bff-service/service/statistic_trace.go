package service

import (
	"context"
	"fmt"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
)

// LookupAppCreator 查询 app 资源创建人（moduleCreator）。
func LookupAppCreator(ctx context.Context, appID, appType string) (userID, orgID string, err error) {
	appInfo, err := app.GetAppInfo(ctx, &app_service.GetAppInfoReq{
		AppId:   appID,
		AppType: appType,
	})
	if err != nil {
		return "", "", err
	}
	return appInfo.UserId, appInfo.OrgId, nil
}

// LookupAppUrlBySuffix 查询 app url 信息。
func LookupAppUrlBySuffix(ctx context.Context, suffix string) (*app_service.AppUrlInfo, error) {
	return app.GetAppUrlInfoBySuffix(ctx, &app_service.GetAppUrlInfoBySuffixReq{Suffix: suffix})
}

// LookupKnowledgeCreator 查询知识库创建人（moduleCreator）。
// 不按调用方 owner 过滤：授权用户也可取到创建人；业务权限由 AuthKnowledge 等中间件负责。
func LookupKnowledgeCreator(ctx context.Context, knowledgeID string) (creatorUserID, creatorOrgID string, err error) {
	resp, err := knowledgeBase.SelectKnowledgeListByIdList(ctx, &knowledgebase_service.BatchKnowledgeSelectReq{
		KnowledgeIdList: []string{knowledgeID},
		NoPermission:    true,
	})
	if err != nil {
		return "", "", err
	}
	list := resp.GetKnowledgeList()
	if len(list) == 0 || list[0] == nil {
		return "", "", fmt.Errorf("knowledge %s not found", knowledgeID)
	}
	return list[0].GetCreateUserId(), list[0].GetCreateOrgId(), nil
}

// LookupModelCreator 查询模型导入人（moduleCreator），与 record 阶段的 modelCreator 一致。
func LookupModelCreator(ctx context.Context, modelID string) (userID, orgID string, err error) {
	info, err := model.GetModel(ctx, &model_service.GetModelReq{ModelId: modelID})
	if err != nil {
		return "", "", err
	}
	return info.GetUserId(), info.GetOrgId(), nil
}
