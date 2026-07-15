package service

import (
	"context"
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
	workFlows, err := explorerationFilterWorkFlow(ctx, explorationApp.Infos, req.Name)
	if err != nil {
		return nil, err
	}
	chatFlows, err := explorerationFilterChatFlow(ctx, explorationApp.Infos, req.Name)
	if err != nil {
		return nil, err
	}
	apps := append(rags, append(agents, append(workFlows, chatFlows...)...)...)
	sort.SliceStable(apps, func(i, j int) bool {
		// 按更新时间降序排序，保证列表刷新时最新的优先展示
		return apps[i].UpdatedAt > apps[j].UpdatedAt
	})
	// 填充作者信息
	var userIds []string
	for _, app := range apps {
		userIds = append(userIds, app.User.UserId)
	}
	ret, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
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
		Total: int64(len(apps)),
	}, nil
}

func GetWorkflowSelect(ctx *gin.Context, userId, orgId string, req request.GetExplorationAppListRequest) (*response.ListResult, error) {
	wlist, err := GetExplorationAppList(ctx, userId, orgId, request.GetExplorationAppListRequest{
		Name:       req.Name,
		AppType:    constant.AppTypeWorkflow,
		SearchType: "all",
	})
	if err != nil {
		return nil, err
	}
	clist, err := GetExplorationAppList(ctx, userId, orgId, request.GetExplorationAppListRequest{
		Name:       req.Name,
		AppType:    constant.AppTypeChatflow,
		SearchType: "all",
	})
	if err != nil {
		return nil, err
	}
	// 类型断言并合并
	var combinedList []*response.ExplorationAppInfo
	// 处理 wlist.List
	if wlistSlice, ok := wlist.List.([]*response.ExplorationAppInfo); ok {
		for _, w := range wlistSlice {
			if w.User.UserId == userId {
				combinedList = append(combinedList, w)
			}
		}
	}

	// 处理 clist.List
	if clistSlice, ok := clist.List.([]*response.ExplorationAppInfo); ok {
		for _, c := range clistSlice {
			if c.User.UserId == userId {
				combinedList = append(combinedList, c)
			}
		}
	}

	return &response.ListResult{
		List:  combinedList,
		Total: int64(len(combinedList)),
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

func AddAppHistoryRecord(ctx context.Context, userId, appId, appType string) error {
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
	// 首先收集所有rag类型的appId
	var ids []string
	for _, info := range explorationApp {
		if info.AppType == constant.AppTypeRag {
			ids = append(ids, info.AppId)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// 获取rag详情
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
					AppBriefInfo: appBriefProto2Model(ctx, foundRag, 0),
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
	// 如果name不为空，过滤结果
	if name != "" {
		var filteredList []*response.ExplorationAppInfo
		for _, ret := range retAppList {
			if strings.Contains(strings.ToLower(ret.Name), strings.ToLower(name)) {
				filteredList = append(filteredList, ret)
			}
		}
		return filteredList, nil
	}
	return retAppList, nil
}

func explorerationFilterAgent(ctx *gin.Context, apps []*app_service.ExplorationAppInfo, name string) ([]*response.ExplorationAppInfo, error) {
	// 首先收集所有agent类型的appId
	var ids []string
	for _, info := range apps {
		if info.AppType == constant.AppTypeAgent {
			ids = append(ids, info.AppId)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// 获取agent详情
	agentList, err := assistant.GetAssistantByIds(ctx.Request.Context(), &assistant_service.GetAssistantByIdsReq{AssistantIdList: ids})
	if err != nil {
		return nil, err
	}
	var retAppList []*response.ExplorationAppInfo
	for _, id := range ids {
		var foundAgent *common.AppBrief
		var category int32
		for _, ragInfo := range agentList.AssistantInfos {
			if ragInfo.GetInfo().AppId == id {
				foundAgent = ragInfo.Info
				category = ragInfo.Category
				break
			}
		}
		if foundAgent == nil {
			continue
		}
		for _, expApp := range apps {
			if expApp.AppId == id {
				appInfo := &response.ExplorationAppInfo{
					AppBriefInfo: appBriefProto2Model(ctx, foundAgent, category),
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
	// 如果name不为空，过滤结果
	if name != "" {
		var filteredList []*response.ExplorationAppInfo
		for _, ret := range retAppList {
			if strings.Contains(strings.ToLower(ret.Name), strings.ToLower(name)) {
				filteredList = append(filteredList, ret)
			}
		}
		return filteredList, nil
	}
	return retAppList, nil
}

func explorerationFilterWorkFlow(ctx *gin.Context, apps []*app_service.ExplorationAppInfo, name string) ([]*response.ExplorationAppInfo, error) {
	// 首先收集所有agent类型的appId
	var ids []string
	for _, info := range apps {
		if info.AppType == constant.AppTypeWorkflow {
			ids = append(ids, info.AppId)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// 获取工作流详情
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

func explorerationFilterChatFlow(ctx *gin.Context, apps []*app_service.ExplorationAppInfo, name string) ([]*response.ExplorationAppInfo, error) {
	// 首先收集所有chatflow类型的appId
	var ids []string
	for _, info := range apps {
		if info.AppType == constant.AppTypeChatflow {
			ids = append(ids, info.AppId)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// 获取工作流详情
	chatFlowList, err := ListWorkflowByIDs(ctx, name, ids)
	if err != nil {
		return nil, err
	}
	var retAppList []*response.ExplorationAppInfo
	for _, id := range ids {
		var foundChatflow *response.CozeWorkflowListDataWorkflow
		for _, workflow := range chatFlowList.Workflows {
			if workflow.WorkflowId == id {
				foundChatflow = workflow
				break
			}
		}
		if foundChatflow == nil {
			continue
		}
		for _, expApp := range apps {
			if expApp.AppId == id {
				appInfo := &response.ExplorationAppInfo{
					AppBriefInfo: cozeChatflowInfo2Model(foundChatflow),
				}
				appInfo.CreatedAt = util.Time2Str(expApp.CreatedAt)
				appInfo.UpdatedAt = util.Time2Str(expApp.UpdatedAt)
				appInfo.PublishType = expApp.PublishType
				appInfo.IsFavorite = expApp.IsFavorite
				appInfo.Avatar = cacheWorkflowAvatar(foundChatflow.URL, constant.AppTypeChatflow)
				retAppList = append(retAppList, appInfo)
				appInfo.User.UserId = expApp.UserId
				break
			}
		}
	}
	return retAppList, nil
}
