package service

import (
	"sort"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	"github.com/gin-gonic/gin"
)

// GetModelStatisticV2Select 模型 Tab 下拉（viewScope 分流：published 主表 / used 聚合表）。
// 返回结构与 V1 一致：ListResult{list=[]*ModelInfo}。
func GetModelStatisticV2Select(ctx *gin.Context, req *request.ModelStatisticV2SelectReq, userId, orgId string, isAdmin, isSystem bool) (*response.ListResult, error) {
	if req.ViewScope == statistic.ViewScopePublished {
		return GetStatisticModelSelect(ctx, req.ModelType, userId, orgId, &req.StatisticFilter, isAdmin, isSystem)
	}
	return getModelStatisticV2SelectUsed(ctx, req, userId, orgId, isAdmin, isSystem)
}

func getModelStatisticV2SelectUsed(ctx *gin.Context, req *request.ModelStatisticV2SelectReq, userId, orgId string, isAdmin, isSystem bool) (*response.ListResult, error) {
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.GetModelStatisticV2Select(ctx.Request.Context(), &app_service.GetModelStatisticV2SelectReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		ModelType: req.ModelType,
		ViewScope: req.ViewScope,
	})
	if err != nil {
		return nil, err
	}

	modelIds := make([]string, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		modelIds = append(modelIds, it.GetModelId())
	}
	modelMap := getModelInfoMap(ctx, modelIds)

	items := make([]*response.ModelInfo, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		info := modelMap[it.GetModelId()]
		displayName := pickModelDisplayName(info, it.GetModel())
		if displayName == "" {
			displayName = gin_util.I18nKey(ctx, "app_statistic_model_deleted")
		}
		modelType := info.ModelType
		if modelType == "" {
			modelType = it.GetModelType()
		}
		items = append(items, &response.ModelInfo{
			ModelId:     it.GetModelId(),
			Model:       it.GetModel(),
			Provider:    it.GetProvider(),
			ModelType:   modelType,
			DisplayName: displayName,
			Avatar:      cacheModelAvatar(ctx, info.modelIconPath),
			UserId:      firstNonEmpty(info.creatorUserId, it.GetModelCreatorUserId()),
			OrgId:       firstNonEmpty(info.creatorOrgId, it.GetModelCreatorOrgId()),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].DisplayName < items[j].DisplayName
	})
	return &response.ListResult{List: items, Total: int64(len(items))}, nil
}

// GetAppStatisticV2Select 应用 Tab 下拉（viewScope 分流：published 主表 / used 聚合表）。
// 支持 module：agent / rag / workflow（含 chatflow，两者一并返回）/ knowledge；其余返回空列表。
// 返回结构与 V1 一致：ListResult{list=[]MyAppItem}。
func GetAppStatisticV2Select(ctx *gin.Context, req *request.AppStatisticV2SelectReq, userId, orgId string, isAdmin, isSystem bool) (*response.ListResult, error) {
	if req.ViewScope == statistic.ViewScopePublished {
		return getAppStatisticV2SelectPublished(ctx, req, userId, orgId, isAdmin, isSystem)
	}
	return getAppStatisticV2SelectUsed(ctx, req, userId, orgId, isAdmin, isSystem)
}

func getAppStatisticV2SelectPublished(ctx *gin.Context, req *request.AppStatisticV2SelectReq, userId, orgId string, isAdmin, isSystem bool) (*response.ListResult, error) {
	module, ok := resolveAppStatisticSelectModule(req.Module)
	if !ok {
		log.Warnf("app statistic v2 select published: module %q not supported, return empty list", req.Module)
		return &response.ListResult{List: []response.MyAppItem{}, Total: 0}, nil
	}
	// knowledge 不走 app 发布表，直接按 scope 调 knowledge-service
	if module == constant.BizModuleResourceKnowledge {
		return getKnowledgeStatisticSelectPublished(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	}
	// 对话流 Trace 写 module=workflow；选「工作流」时合并已发布的 workflow + chatflow
	if module == constant.BizModuleAppWorkflow {
		workflowList, err := GetAppListSelect(ctx, req.StatisticFilter, constant.AppTypeWorkflow, userId, orgId, isAdmin, isSystem)
		if err != nil {
			return nil, err
		}
		chatflowList, err := GetAppListSelect(ctx, req.StatisticFilter, constant.AppTypeChatflow, userId, orgId, isAdmin, isSystem)
		if err != nil {
			return nil, err
		}
		items := append(workflowList.List.([]response.MyAppItem), chatflowList.List.([]response.MyAppItem)...)
		sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
		return &response.ListResult{List: items, Total: int64(len(items))}, nil
	}
	appType, ok := statisticModuleToPublishedAppType(module)
	if !ok {
		log.Warnf("app statistic v2 select published: module %q not supported, return empty list", req.Module)
		return &response.ListResult{List: []response.MyAppItem{}, Total: 0}, nil
	}
	return GetAppListSelect(ctx, req.StatisticFilter, appType, userId, orgId, isAdmin, isSystem)
}

func getKnowledgeStatisticSelectPublished(ctx *gin.Context, filter request.StatisticFilter, userId, orgId string, isAdmin, isSystem bool) (*response.ListResult, error) {
	scope, err := ResolveStatisticScope(ctx, filter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := knowledgeBase.SelectKnowledgeListByUserList(ctx.Request.Context(), &knowledgebase_service.KnowledgeSelectUserListReq{
		UserId: scope.UserIds,
		OrgId:  scope.OrgIds,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeMyAppItems(ctx, resp.GetKnowledgeList()), nil
}

func getAppStatisticV2SelectUsed(ctx *gin.Context, req *request.AppStatisticV2SelectReq, userId, orgId string, isAdmin, isSystem bool) (*response.ListResult, error) {
	module, ok := resolveAppStatisticSelectModule(req.Module)
	if !ok {
		log.Warnf("app statistic v2 select used: module %q not supported, return empty list", req.Module)
		return &response.ListResult{List: []response.MyAppItem{}, Total: 0}, nil
	}
	scope, err := ResolveStatisticScope(ctx, req.StatisticFilter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	// used：workflow 查 module=workflow 且 app_type∈{workflow,chatflow}；agent/rag/knowledge 查对应类型（ORM 内落地）
	resp, err := app.GetAppStatisticV2Select(ctx.Request.Context(), &app_service.GetAppStatisticV2SelectReq{
		OrgIds:    scope.OrgIds,
		UserIds:   scope.UserIds,
		Module:    module,
		ViewScope: req.ViewScope,
	})
	if err != nil {
		return nil, err
	}

	// knowledge 写入统计时 appType 可能为空，按 module 走 id 批量补全
	if module == constant.BizModuleResourceKnowledge {
		return getKnowledgeStatisticSelectUsed(ctx, resp.GetItems())
	}

	appIdsByType := map[string]map[string]struct{}{}
	for _, it := range resp.GetItems() {
		if it.GetAppId() == "" || it.GetAppType() == "" {
			continue
		}
		m, ok := appIdsByType[it.GetAppType()]
		if !ok {
			m = map[string]struct{}{}
			appIdsByType[it.GetAppType()] = m
		}
		m[it.GetAppId()] = struct{}{}
	}
	appBriefMap := getAppBriefMapMulti(ctx, appIdsByType)

	items := make([]response.MyAppItem, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		if it.GetAppId() == "" || it.GetAppType() == "" {
			continue
		}
		items = append(items, response.MyAppItem{
			AppId:   it.GetAppId(),
			Name:    pickAppName(appBriefMap, it.GetAppType(), it.GetAppId(), gin_util.I18nKey(ctx, "app_statistic_app_deleted")),
			AppType: it.GetAppType(),
			Avatar:  pickAppAvatar(appBriefMap, it.GetAppType(), it.GetAppId()),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return &response.ListResult{List: items, Total: int64(len(items))}, nil
}

func getKnowledgeStatisticSelectUsed(ctx *gin.Context, statItems []*app_service.AppStatisticV2SelectItem) (*response.ListResult, error) {
	knowledgeIds := make([]string, 0, len(statItems))
	seen := make(map[string]struct{}, len(statItems))
	for _, it := range statItems {
		id := it.GetAppId()
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		knowledgeIds = append(knowledgeIds, id)
	}
	if len(knowledgeIds) == 0 {
		return &response.ListResult{List: []response.MyAppItem{}, Total: 0}, nil
	}
	resp, err := knowledgeBase.SelectKnowledgeListByIdList(ctx.Request.Context(), &knowledgebase_service.BatchKnowledgeSelectReq{
		KnowledgeIdList: knowledgeIds,
		NoPermission:    true,
	})
	if err != nil {
		return nil, err
	}
	infoMap := make(map[string]*knowledgebase_service.KnowledgeInfo, len(resp.GetKnowledgeList()))
	for _, k := range resp.GetKnowledgeList() {
		if k != nil && k.GetKnowledgeId() != "" {
			infoMap[k.GetKnowledgeId()] = k
		}
	}
	items := make([]response.MyAppItem, 0, len(knowledgeIds))
	for _, id := range knowledgeIds {
		item := response.MyAppItem{
			AppId:   id,
			Name:    gin_util.I18nKey(ctx, "app_statistic_app_deleted"),
			AppType: constant.BizModuleResourceKnowledge,
		}
		if k, ok := infoMap[id]; ok {
			item.Name = k.GetName()
			item.Avatar = cacheKnowledgeAvatar(ctx, k.GetAvatarPath(), k.GetCategory())
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return &response.ListResult{List: items, Total: int64(len(items))}, nil
}

func buildKnowledgeMyAppItems(ctx *gin.Context, list []*knowledgebase_service.KnowledgeInfo) *response.ListResult {
	items := make([]response.MyAppItem, 0, len(list))
	for _, k := range list {
		if k == nil || k.GetKnowledgeId() == "" {
			continue
		}
		items = append(items, response.MyAppItem{
			AppId:   k.GetKnowledgeId(),
			Name:    k.GetName(),
			AppType: constant.BizModuleResourceKnowledge,
			Avatar:  cacheKnowledgeAvatar(ctx, k.GetAvatarPath(), k.GetCategory()),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return &response.ListResult{List: items, Total: int64(len(items))}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// resolveAppStatisticSelectModule 下拉 module 归一化。
// chatflow 写入口径为 module=workflow，请求 chatflow 时按 workflow 处理（返回 workflow+chatflow）。
func resolveAppStatisticSelectModule(module string) (string, bool) {
	switch module {
	case constant.BizModuleAppAgent, constant.BizModuleAppRag, constant.BizModuleAppWorkflow, constant.BizModuleResourceKnowledge:
		return module, true
	case constant.AppTypeChatflow:
		return constant.BizModuleAppWorkflow, true
	default:
		return "", false
	}
}

// statisticModuleToPublishedAppType 将 V2 module 映射为发布表 appType。
// 仅 agent / rag；workflow 合并 chatflow、knowledge 走 knowledge-service，均在调用方单独处理。
func statisticModuleToPublishedAppType(module string) (appType string, ok bool) {
	switch module {
	case constant.BizModuleAppAgent:
		return constant.AppTypeAgent, true
	case constant.BizModuleAppRag:
		return constant.AppTypeRag, true
	default:
		return "", false
	}
}
