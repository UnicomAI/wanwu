package service

import (
	"sort"
	"strings"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func GetExplorationAppList(ctx *gin.Context, userId, orgId string, req request.GetExplorationAppListRequest) (*response.ListResult, error) {
	explorationApp, err := app.GetExplorationAppList(ctx.Request.Context(), &app_service.GetExplorationAppListReq{
		Name:       req.Name,
		AppType:    req.AppType,
		SearchType: req.SearchType,
		UserId:     userId,
		OrgId:      orgId,
	})
	if err != nil {
		return nil, err
	}
	rags, err := explorerationFilterRag(ctx, explorationApp.Infos, req.Name)
	if err != nil {
		return nil, err
	}
	agents, err := explorerationFilterAgent(ctx, explorationApp.Infos, req.Name)
	if err != nil {
		return nil, err
	}
	// AgentScope Workflow
	// workFlows, err := explorerationFilterAgentScopeWorkFlow(ctx, explorationApp.Infos, req.Name)
	// Coze Workflow
	workFlows, err := explorerationFilterWorkFlow(ctx, explorationApp.Infos, req.Name)
	if err != nil {
		return nil, err
	}
	apps := append(rags, append(agents, workFlows...)...)
	sort.SliceStable(apps, func(i, j int) bool {
		return apps[i].CreatedAt > apps[j].CreatedAt
	})
	// 填充作者信息 [EN] Fill in author information
	var userIds []string
	for _, app := range apps {
		userIds = append(userIds, app.User.UserId)
	}
	ret, err := iam.GetUserSelectByUserIDs(ctx, &iam_service.GetUserSelectByUserIDsReq{
		UserIds: userIds,
	})
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		app.User.UserName = gin_util.I18nKey(ctx, "iam_user_deleted")
		for _, user := range ret.Selects {
			if app.User.UserId == user.Id {
				app.User.UserName = user.Name
			}
		}
	}
	return &response.ListResult{
		List:  apps,
		Total: explorationApp.Total,
	}, nil
}

func ChangeExplorationAppFavorite(ctx *gin.Context, userId, orgId string, req request.ChangeExplorationAppFavoriteRequest) error {
	_, err := app.ChangeExplorationAppFavorite(ctx.Request.Context(), &app_service.ChangeExplorationAppFavoriteReq{
		AppId:      req.AppId,
		AppType:    req.AppType,
		UserId:     userId,
		OrgId:      orgId,
		IsFavorite: req.IsFavorite,
	})
	return err
}

func AddAppHistoryRecord(ctx *gin.Context, userId, appId, appType string) error {
	if _, err := app.RecordAppHistory(ctx, &app_service.RecordAppHistoryReq{
		UserId:  userId,
		AppId:   appId,
		AppType: appType,
	}); err != nil {
		return err
	}
	return nil

}

// --- internal ---

func explorerationFilterRag(ctx *gin.Context, explorationApp []*app_service.ExplorationAppInfo, name string) ([]*response.ExplorationAppInfo, error) {
	// 首先收集所有rag类型的appId [EN] First collect the appIds of all rag types
	var ids []string
	for _, info := range explorationApp {
		if info.AppType == constant.AppTypeRag {
			ids = append(ids, info.AppId)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// 获取rag详情 [EN] Get rag details
	ragList, err := rag.GetRagByIds(ctx.Request.Context(), &rag_service.GetRagByIdsReq{RagIdList: ids})
	if err != nil {
		return nil, err
	}
	var retAppList []*response.ExplorationAppInfo
	for _, id := range ids {
		var foundRag *common.AppBrief
		for _, ragInfo := range ragList.RagInfos {
			if ragInfo.AppId == id {
				foundRag = ragInfo
				break
			}
		}
		if foundRag == nil {
			continue
		}
		for _, expApp := range explorationApp {
			if expApp.AppId == id {
				appInfo := &response.ExplorationAppInfo{
					AppBriefInfo: appBriefProto2Model(ctx, foundRag),
				}
				appInfo.CreatedAt = util.Time2Str(expApp.CreatedAt)
				appInfo.UpdatedAt = util.Time2Str(expApp.UpdatedAt)
				appInfo.PublishType = expApp.PublishType
				appInfo.IsFavorite = expApp.IsFavorite
				appInfo.User.UserId = expApp.UserId
				retAppList = append(retAppList, appInfo)
				break
			}
		}
	}
	// 如果name不为空，过滤结果 [EN] If name is not empty, filter the results
	if name != "" {
		var filteredList []*response.ExplorationAppInfo
		for _, ret := range retAppList {
			if strings.Contains(strings.ToLower(ret.AppBriefInfo.Name), strings.ToLower(name)) {
				filteredList = append(filteredList, ret)
			}
		}
		return filteredList, nil
	}
	return retAppList, nil
}

func explorerationFilterAgent(ctx *gin.Context, apps []*app_service.ExplorationAppInfo, name string) ([]*response.ExplorationAppInfo, error) {
	// 首先收集所有agent类型的appId [EN] First collect the appIds of all agent types
	var ids []string
	for _, info := range apps {
		if info.AppType == constant.AppTypeAgent {
			ids = append(ids, info.AppId)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// 获取agent详情 [EN] Get agent details
	agentList, err := assistant.GetAssistantByIds(ctx.Request.Context(), &assistant_service.GetAssistantByIdsReq{AssistantIdList: ids})
	if err != nil {
		return nil, err
	}
	var retAppList []*response.ExplorationAppInfo
	for _, id := range ids {
		var foundAgent *common.AppBrief
		for _, ragInfo := range agentList.AssistantInfos {
			if ragInfo.AppId == id {
				foundAgent = ragInfo
				break
			}
		}
		if foundAgent == nil {
			continue
		}
		for _, expApp := range apps {
			if expApp.AppId == id {
				appInfo := &response.ExplorationAppInfo{
					AppBriefInfo: appBriefProto2Model(ctx, foundAgent),
				}
				appInfo.CreatedAt = util.Time2Str(expApp.CreatedAt)
				appInfo.UpdatedAt = util.Time2Str(expApp.UpdatedAt)
				appInfo.PublishType = expApp.PublishType
				appInfo.IsFavorite = expApp.IsFavorite
				appInfo.User.UserId = expApp.UserId
				retAppList = append(retAppList, appInfo)
				break
			}
		}
	}
	// 如果name不为空，过滤结果 [EN] If name is not empty, filter the results
	if name != "" {
		var filteredList []*response.ExplorationAppInfo
		for _, ret := range retAppList {
			if strings.Contains(strings.ToLower(ret.AppBriefInfo.Name), strings.ToLower(name)) {
				filteredList = append(filteredList, ret)
			}
		}
		return filteredList, nil
	}
	return retAppList, nil
}

func explorerationFilterWorkFlow(ctx *gin.Context, apps []*app_service.ExplorationAppInfo, name string) ([]*response.ExplorationAppInfo, error) {
	// 首先收集所有agent类型的appId [EN] First collect the appIds of all agent types
	var ids []string
	for _, info := range apps {
		if info.AppType == constant.AppTypeWorkflow {
			ids = append(ids, info.AppId)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// 获取工作流详情 [EN] Get workflow details
	workFlowList, err := ListWorkflowByIDs(ctx, name, ids)
	if err != nil {
		return nil, err
	}
	var retAppList []*response.ExplorationAppInfo
	for _, id := range ids {
		var foundWorkflow *response.CozeWorkflowListDataWorkflow
		for _, workflow := range workFlowList.Workflows {
			if workflow.WorkflowId == id {
				foundWorkflow = workflow
				break
			}
		}
		if foundWorkflow == nil {
			continue
		}
		for _, expApp := range apps {
			if expApp.AppId == id {
				appInfo := &response.ExplorationAppInfo{
					AppBriefInfo: cozeWorkflowInfo2Model(foundWorkflow),
				}
				appInfo.CreatedAt = util.Time2Str(expApp.CreatedAt)
				appInfo.UpdatedAt = util.Time2Str(expApp.UpdatedAt)
				appInfo.PublishType = expApp.PublishType
				appInfo.IsFavorite = expApp.IsFavorite
				appInfo.Avatar = cacheWorkflowAvatar(foundWorkflow.URL, constant.AppTypeWorkflow)
				retAppList = append(retAppList, appInfo)
				appInfo.User.UserId = expApp.UserId
				break
			}
		}
	}
	return retAppList, nil
}

// func explorerationFilterAgentScopeWorkFlow(ctx *gin.Context, apps []*app_service.ExplorationAppInfo, name string) ([]*response.ExplorationAppInfo, error) {
// 	// 获取工作流详情 [EN] // Get workflow details
// 	workFlowList, err := ListAgentScopeWorkFlowInternal(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var retAppList []*response.ExplorationAppInfo
// 	for _, expApp := range apps {
// 		for _, workFlow := range workFlowList.List {
// 			if expApp.AppId == workFlow.Id {
// 				appInfo := &response.ExplorationAppInfo{
// 					AppBriefInfo: response.AppBriefInfo{
// 						AppId:   workFlow.Id,
// 						AppType: constant.AppTypeWorkflow,
// 						Avatar:  request.Avatar{},
// 						Name:    workFlow.ConfigName,
// 						Desc:    workFlow.ConfigDesc,
// 					},
// 				}
// 				appInfo.CreatedAt = util.Time2Str(expApp.CreatedAt)
// 				appInfo.UpdatedAt = util.Time2Str(expApp.UpdatedAt)
// 				appInfo.PublishType = expApp.PublishType
// 				appInfo.IsFavorite = expApp.IsFavorite
// 				retAppList = append(retAppList, appInfo)
// 				break
// 			}
// 		}
// 	}
// 	// 如果name不为空，过滤结果 [EN] // If name is not empty, filter the results
// 	if name != "" {
// 		var filteredList []*response.ExplorationAppInfo
// 		for _, ret := range retAppList {
// 			if strings.Contains(strings.ToLower(ret.AppBriefInfo.Name), strings.ToLower(name)) {
// 				filteredList = append(filteredList, ret)
// 			}
// 		}
// 		return filteredList, nil
// 	}
// 	return retAppList, nil
// }
