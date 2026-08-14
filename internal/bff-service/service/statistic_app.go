package service

import (
	"context"
	"fmt"
	"sort"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/gin-gonic/gin"
)

// appBrief 应用补全信息（name + avatar，同一次 RPC 解析）。
type appBrief struct {
	Name   string
	Avatar request.Avatar
}

// getAppBriefMap 按 appType 一次 RPC 批量拉取 name + avatar（同一批接口响应里两者都有）。
func getAppBriefMap(ctx *gin.Context, appId []string, appType string) (map[string]appBrief, error) {
	result := make(map[string]appBrief)
	switch appType {
	case constant.AppTypeAgent:
		agentListInfos, err := assistant.GetAssistantByIds(ctx.Request.Context(), &assistant_service.GetAssistantByIdsReq{AssistantIdList: appId})
		if err != nil {
			log.Errorf("get agent info err: %v", err)
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("get app brief error: %v", err))
		}
		for _, info := range agentListInfos.AssistantInfos {
			if info.Info == nil {
				continue
			}
			result[info.Info.AppId] = appBrief{
				Name:   info.Info.Name,
				Avatar: cacheAppAvatar(ctx, info.Info.AvatarPath, appType),
			}
		}
		return result, nil
	case constant.AppTypeRag:
		ragListInfos, err := rag.GetRagByIds(ctx.Request.Context(), &rag_service.GetRagByIdsReq{RagIdList: appId})
		if err != nil {
			log.Errorf("get rag info err: %v", err)
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("get app brief error: %v", err))
		}
		for _, info := range ragListInfos.RagInfos {
			result[info.AppId] = appBrief{
				Name:   info.Name,
				Avatar: cacheAppAvatar(ctx, info.AvatarPath, appType),
			}
		}
		return result, nil
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		workflowRet, err := ListWorkflowByIDs(ctx, "", appId)
		if err != nil {
			log.Errorf("get workflow info err: %v", err)
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("get app brief error: %v", err))
		}
		for _, info := range workflowRet.Workflows {
			result[info.WorkflowId] = appBrief{
				Name:   info.Name,
				Avatar: cacheWorkflowAvatar(info.URL, appType),
			}
		}
		return result, nil
	case constant.AppTypeDigitalEmployee:
		// 数字员工名称：复用批量 latest 接口（GetDigitalEmployeeBatchInfo → POST /versions/latest，dh_ids）
		// 契约 latest 无头像字段，用默认图标
		infos, err := GetDigitalEmployeeBatchInfo(ctx, getUserID(ctx), getOrgID(ctx), appId)
		if err != nil {
			log.Errorf("get digital employee info err: %v", err)
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("get app brief error: %v", err))
		}
		for dhID, v := range infos {
			result[dhID] = appBrief{
				Name:   v.Name,
				Avatar: request.Avatar{Path: "/v1/static/icon/wga-digital-employee-icon.svg"},
			}
		}
		return result, nil
	case constant.BizModuleResourceKnowledge:
		// 知识库不在 app 发布表，走 knowledge-service 批量补全名称/头像
		kbResp, err := knowledgeBase.SelectKnowledgeListByIdList(ctx.Request.Context(), &knowledgebase_service.BatchKnowledgeSelectReq{
			KnowledgeIdList: appId,
			NoPermission:    true,
		})
		if err != nil {
			log.Errorf("get knowledge info err: %v", err)
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("get app brief error: %v", err))
		}
		for _, info := range kbResp.GetKnowledgeList() {
			if info == nil || info.GetKnowledgeId() == "" {
				continue
			}
			result[info.GetKnowledgeId()] = appBrief{
				Name:   info.GetName(),
				Avatar: cacheKnowledgeAvatar(ctx, info.GetAvatarPath(), info.GetCategory()),
			}
		}
		return result, nil
	}
	return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("unsupported app type: %v", appType))
}

// getAppBriefMapMulti 按 appType 分组一次拉 name+avatar：appType → appId → brief。
func getAppBriefMapMulti(ctx *gin.Context, appIdsByType map[string]map[string]struct{}) map[string]map[string]appBrief {
	merged := map[string]map[string]appBrief{}
	for appType, idsSet := range appIdsByType {
		if len(idsSet) == 0 {
			continue
		}
		ids := make([]string, 0, len(idsSet))
		for id := range idsSet {
			ids = append(ids, id)
		}
		briefs, err := getAppBriefMap(ctx, ids, appType)
		if err != nil {
			log.Warnf("getAppBriefMapMulti appType %v err: %v, appIds: %v", appType, err, ids)
			continue
		}
		merged[appType] = briefs
	}
	return merged
}

func pickAppBrief(appBriefMap map[string]map[string]appBrief, appType, appId string) (appBrief, bool) {
	if appBriefMap == nil || appType == "" || appId == "" {
		return appBrief{}, false
	}
	byType, ok := appBriefMap[appType]
	if !ok {
		return appBrief{}, false
	}
	brief, ok := byType[appId]
	return brief, ok
}

func pickAppName(appBriefMap map[string]map[string]appBrief, appType, appId, fallback string) string {
	if brief, ok := pickAppBrief(appBriefMap, appType, appId); ok && brief.Name != "" {
		return brief.Name
	}
	return fallback
}

// pickAppAvatar 应用仍在时用补全头像；已删除/查不到时按 appType 复用对应 cache*Avatar
// 传空 path 拿默认头像 path（agent/rag→cacheAppAvatar，workflow/chatflow→cacheWorkflowAvatar，
// knowledge→cacheKnowledgeAvatar）。
func pickAppAvatar(ctx *gin.Context, appBriefMap map[string]map[string]appBrief, appType, appId string) request.Avatar {
	if brief, ok := pickAppBrief(appBriefMap, appType, appId); ok {
		return brief.Avatar
	}
	switch appType {
	case constant.AppTypeAgent, constant.AppTypeRag:
		return cacheAppAvatar(ctx, "", appType)
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		return cacheWorkflowAvatar("", appType)
	case constant.BizModuleResourceKnowledge:
		return cacheKnowledgeAvatar(ctx, "", constant.KnowledgeBase)
	}
	return request.Avatar{}
}

// RecordAppStatistic 记录一次应用调用统计（V2）。
// 草稿对话也会写入（与 V1「草稿不统计」不同，属 V2 产品约定）。
// statusCode/failureReason 为该次调用的真实结果：调用方应优先从实际 HTTP/gRPC
// 错误中提取（参见 grpcErrorToHTTPStatus）；仅有 isSuccess bool 时，可用 200/500
// 与 "会话执行失败" 兜底。isSuccess 由 statusCode 内部反推，不再单独传入。
// requestBody/responseBody 应为整体 req/resp 的 JSON（MarshalStatisticBody）；受 statistic.record_body 控制；流式仅落 requestBody。
// question/answer 为精简问答摘要，始终落库，不受 statistic.record_body 控制；流式 answer 底层强制清空。
// module 可显式传入（如 callback/knowledge）；为空则按 Trace > appType 映射解析，仍空则跳过写入。
// appType 解析优先级：显式传入 > Trace.moduleResourceType（与模型统计一致）。
func RecordAppStatistic(ctx context.Context, userId, orgId, appId, appType, module string,
	statusCode int64, failureReason string,
	isStream bool, streamCosts, nonStreamCosts int64, source, requestBody, responseBody, question, answer string) {
	// 同一次落库只 Parse 一次，避免 resolve appType/module/creator 各打一遍 Redis。
	statCtx := parseStatisticContextOptional(ctx)
	if appType == "" {
		appType = resolveAppStatisticAppType(statCtx)
	}
	if appId == "" && statCtx != nil {
		// 从 Trace 补 appId（如数字员工发布对话 TraceWeb 设了 employeeId；
		// wga 通用管线传空 appId，靠这里补上具体应用，避免应用级模块空 appId 被拦截）
		appId = statCtx.AppID
	}
	if module == "" {
		module = resolveAppStatisticModule(statCtx, appType)
	}
	if module == "" {
		log.Errorf("record app statistic skip: module unresolved, appId=%v appType=%v source=%v", appId, appType, source)
		return
	}
	moduleCreatorUserId, moduleCreatorOrgId := resolveAppStatisticModuleCreator(ctx, statCtx, userId, orgId, appId, appType)
	recordAppStatisticV2(ctx, userId, orgId, appId, appType, module, moduleCreatorUserId, moduleCreatorOrgId,
		statusCode, failureReason, isStream, streamCosts, nonStreamCosts, source, requestBody, responseBody, question, answer)
}

func parseStatisticContextOptional(ctx context.Context) *trace_util.StatisticContext {
	statCtx, err := trace_util.ParseStatisticContext(ctx)
	if err != nil {
		return nil
	}
	return statCtx
}

// resolveAppStatisticAppType 从已解析的 Trace 读取 appType（moduleResourceType）。
func resolveAppStatisticAppType(statCtx *trace_util.StatisticContext) string {
	if statCtx == nil {
		return ""
	}
	return statCtx.AppType
}

// resolveAppStatisticModuleCreator 优先 Trace 中的 moduleCreator；其次已发布 app 创建人；草稿/资源类回退调用人。
func resolveAppStatisticModuleCreator(ctx context.Context, statCtx *trace_util.StatisticContext,
	callerUserId, callerOrgId, appId, appType string) (userID, orgID string) {
	if statCtx != nil && statCtx.ModuleCreatorUserID != "" && statCtx.ModuleCreatorOrgID != "" {
		return statCtx.ModuleCreatorUserID, statCtx.ModuleCreatorOrgID
	}
	if appId == "" || appType == "" {
		return callerUserId, callerOrgId
	}
	appInfo, err := app.GetAppInfo(ctx, &app_service.GetAppInfoReq{
		AppId:   appId,
		AppType: appType,
	})
	if err == nil {
		return appInfo.UserId, appInfo.OrgId
	}
	return callerUserId, callerOrgId
}

func recordAppStatisticV2(ctx context.Context, callerUserId, callerOrgId, appId, appType, module,
	moduleCreatorUserId, moduleCreatorOrgId string,
	statusCode int64, failureReason string,
	isStream bool, streamCosts, nonStreamCosts int64, source, requestBody, responseBody, question, answer string) {
	costs := nonStreamCosts
	ftl := int64(0)
	if isStream {
		ftl = streamCosts
		costs = 0
		responseBody = "" // 流式不落 ResponseBody
		answer = ""       // 流式不落 Answer
	}
	isSuccess := statistic.IsSuccess(statusCode)
	req := &app_service.RecordAppStatisticV2Req{
		UserId:              callerUserId,
		OrgId:               callerOrgId,
		Source:              source,
		Module:              module,
		AppId:               appId,
		AppType:             appType,
		ModuleCreatorUserId: moduleCreatorUserId,
		ModuleCreatorOrgId:  moduleCreatorOrgId,
		IsSuccess:           isSuccess,
		IsStream:            isStream,
		FirstTokenLatency:   ftl,
		Costs:               costs,
		StatusCode:          statusCode,
		RequestBody:         maybeRecordBody(requestBody),
		ResponseBody:        maybeRecordBody(responseBody),
		Question:            question,
		Answer:              answer,
		FailureReason:       failureReason,
	}
	if trace_util.IsTraceContextValid(ctx) {
		req.TraceId = trace_util.GetTraceID(ctx)
	}
	_, err := app.RecordAppStatisticV2(ctx, req)
	if err != nil {
		log.Errorf("record app v2 %v type %v module %v source %v err: %v", appId, appType, module, source, err)
	}
}

// resolveAppStatisticModule 在未显式传 module 时解析：Trace > appType 映射。
func resolveAppStatisticModule(statCtx *trace_util.StatisticContext, appType string) string {
	if statCtx != nil && statCtx.Module != "" {
		return statCtx.Module
	}
	switch appType {
	case constant.AppTypeAgent:
		return constant.BizModuleAppAgent
	case constant.AppTypeRag:
		return constant.BizModuleAppRag
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		return constant.BizModuleAppWorkflow
	case constant.AppTypeDigitalEmployee:
		return constant.BizModuleAppDigitalEmployee
	default:
		return ""
	}
}

// resolveAppStatisticSource 优先 Trace source，缺失时用 fallback。
func resolveAppStatisticSource(ctx context.Context, fallback string) string {
	statCtx := parseStatisticContextOptional(ctx)
	if statCtx != nil && statCtx.Source != "" {
		return statCtx.Source
	}
	if fallback != "" {
		return fallback
	}
	return constant.BizSourceWeb
}

func GetAppListSelect(ctx *gin.Context, filter request.StatisticFilter, appType string, userId, orgId string, isAdmin, isSystem bool) (*response.ListResult, error) {
	scope, err := ResolveStatisticScope(ctx, filter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	if appType == "" {
		appType = constant.AppTypeAgent
	}
	resp, err := app.GetAppList(ctx.Request.Context(), &app_service.GetAppListReq{
		OrgIds:  scope.OrgIds,
		UserIds: scope.UserIds,
		AppType: appType,
	})
	if err != nil {
		return nil, err
	}

	var appIds []string
	for _, info := range resp.Infos {
		appIds = append(appIds, info.AppId)
	}

	items := make([]response.MyAppItem, 0, len(resp.Infos))

	switch appType {
	case constant.AppTypeAgent:
		agentInfos, err := assistant.GetAssistantByIds(ctx.Request.Context(), &assistant_service.GetAssistantByIdsReq{
			AssistantIdList: appIds,
		})
		if err != nil {
			log.Errorf("app select get agent info err: %v", err)
			return nil, err
		}
		agentMap := make(map[string]*common.AppBrief)
		for _, info := range agentInfos.AssistantInfos {
			if info.Info != nil {
				agentMap[info.Info.AppId] = info.Info
			}
		}
		for _, info := range resp.Infos {
			item := response.MyAppItem{
				AppId:       info.AppId,
				AppType:     info.AppType,
				PublishType: info.PublishType,
				CreatedAt:   info.CreatedAt,
			}
			if agentInfo, ok := agentMap[info.AppId]; ok {
				item.Name = agentInfo.Name
				item.Avatar = cacheAppAvatar(ctx, agentInfo.AvatarPath, appType)
			}
			items = append(items, item)
		}

	case constant.AppTypeRag:
		ragInfos, err := rag.GetRagByIds(ctx.Request.Context(), &rag_service.GetRagByIdsReq{
			RagIdList: appIds,
		})
		if err != nil {
			log.Errorf("app select get rag info err: %v", err)
			return nil, err
		}
		ragMap := make(map[string]*common.AppBrief)
		for _, info := range ragInfos.RagInfos {
			ragMap[info.AppId] = info
		}
		for _, info := range resp.Infos {
			item := response.MyAppItem{
				AppId:       info.AppId,
				AppType:     info.AppType,
				PublishType: info.PublishType,
				CreatedAt:   info.CreatedAt,
			}
			if ragInfo, ok := ragMap[info.AppId]; ok {
				item.Name = ragInfo.Name
				item.Avatar = cacheAppAvatar(ctx, ragInfo.AvatarPath, appType)
			}
			items = append(items, item)
		}

	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		workflowRet, err := ListWorkflowByIDs(ctx, "", appIds)
		if err != nil {
			log.Errorf("app select get workflow info err: %v", err)
			return nil, err
		}
		workflowMap := make(map[string]*response.CozeWorkflowListDataWorkflow)
		for _, info := range workflowRet.Workflows {
			workflowMap[info.WorkflowId] = info
		}
		for _, info := range resp.Infos {
			item := response.MyAppItem{
				AppId:       info.AppId,
				AppType:     info.AppType,
				PublishType: info.PublishType,
				CreatedAt:   info.CreatedAt,
			}
			if workflowInfo, ok := workflowMap[info.AppId]; ok {
				item.Name = workflowInfo.Name
				item.Avatar = cacheWorkflowAvatar(workflowInfo.URL, appType)
			}
			items = append(items, item)
		}

	case constant.AppTypeDigitalEmployee:
		// 数字员工名称：复用批量 latest（GetDigitalEmployeeBatchInfo → POST /versions/latest，dh_ids）
		// 契约 latest 无头像字段，用默认图标
		deInfos, err := GetDigitalEmployeeBatchInfo(ctx, getUserID(ctx), getOrgID(ctx), appIds)
		if err != nil {
			log.Errorf("app select get digital employee info err: %v", err)
			return nil, err
		}
		for _, info := range resp.Infos {
			item := response.MyAppItem{
				AppId:       info.AppId,
				AppType:     info.AppType,
				PublishType: info.PublishType,
				CreatedAt:   info.CreatedAt,
			}
			if deInfo, ok := deInfos[info.AppId]; ok {
				item.Name = deInfo.Name
				item.Avatar = request.Avatar{Path: "/v1/static/icon/wga-digital-employee-icon.svg"}
			}
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return &response.ListResult{
		List:  items,
		Total: int64(len(items)),
	}, nil
}
