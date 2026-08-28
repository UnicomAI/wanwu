package service

import (
	"context"
	"fmt"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/gin-gonic/gin"
)

type apiKeyInfo struct {
	name string
	key  string
}

func GetStatisticAPIKeySelect(ctx *gin.Context, filter request.StatisticFilter, userId, orgId string, isAdmin, isSystem bool) (*response.ListResult, error) {
	scope, err := ResolveStatisticScope(ctx, filter, userId, orgId, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}
	resp, err := app.ListApiKeys(ctx.Request.Context(), &app_service.ListApiKeysReq{
		OrgIds:   scope.OrgIds,
		UserIds:  scope.UserIds,
		PageNo:   1,
		PageSize: 1000,
	})
	if err != nil {
		return nil, err
	}
	items := make([]response.APIKeyDetailResponse, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, response.APIKeyDetailResponse{
			KeyID: item.KeyId,
			Name:  item.Name,
			Key:   item.Key,
		})
	}
	return &response.ListResult{
		List: items,
	}, nil
}

func RecordAPIKeyCall(ctx context.Context, userId, orgId, apiKeyId, methodPath string,
	callTime int64, statusCode int64, isStream bool, streamCosts, nonStreamCosts int64, requestBody, responseBody string) {
	isSuccess := statistic.IsSuccess(statusCode)
	costs := nonStreamCosts
	ftl := int64(0)
	if isStream {
		ftl = streamCosts
		costs = 0
	}
	req := &app_service.RecordAPIKeyStatisticV2Req{
		UserId:            userId,
		OrgId:             orgId,
		ApiKeyId:          apiKeyId,
		MethodPath:        methodPath,
		CalledAt:          callTime,
		Source:            constant.BizSourceOpenAPI,
		IsStream:          isStream,
		Costs:             costs,
		FirstTokenLatency: ftl,
		IsSuccess:         isSuccess,
		StatusCode:        statusCode,
		RequestBody:       maybeRecordBody(requestBody),
		ResponseBody:      maybeRecordBody(responseBody),
	}
	if !isSuccess {
		req.FailureReason = fmt.Sprintf("HTTP %d", statusCode)
	}
	if statCtx, err := trace_util.ParseStatisticContext(ctx); err == nil {
		req.Source = statCtx.Source
		req.Module = statCtx.Module
		req.ModuleCreatorUserId = statCtx.ModuleCreatorUserID
		req.ModuleCreatorOrgId = statCtx.ModuleCreatorOrgID
		req.AppId = statCtx.AppID
		req.AppType = statCtx.AppType
		req.TraceId = statCtx.TraceID
	} else if trace_util.IsTraceContextValid(ctx) {
		req.TraceId = trace_util.GetTraceID(ctx)
	}
	_, err := app.RecordAPIKeyStatisticV2(ctx, req)
	if err != nil {
		log.Errorf("record api key v2[%v] method_path[%v] call err: %v", apiKeyId, methodPath, err)
	}
}

// normalizeAPIKeyIds apiKeyIds 为 ["ALL"] 时不按 key 过滤（与 StatisticFilter 的 ALL 无关）。
func normalizeAPIKeyIds(ids []string) []string {
	if len(ids) == 1 && ids[0] == "ALL" {
		return nil
	}
	return ids
}

func getAPIKeyInfoMap(ctx *gin.Context, scope *statisticScope) map[string]apiKeyInfo {
	resp, err := app.ListApiKeys(ctx.Request.Context(), &app_service.ListApiKeysReq{
		OrgIds:   scope.OrgIds,
		UserIds:  scope.UserIds,
		PageNo:   -1,
		PageSize: -1,
	})
	if err != nil {
		log.Warnf("get api key info map err: %v", err)
		return nil
	}
	infoMap := make(map[string]apiKeyInfo)
	for _, item := range resp.Items {
		infoMap[item.KeyId] = apiKeyInfo{
			name: item.Name,
			key:  item.Key,
		}
	}
	return infoMap
}

func getAPIKeyDisplayInfo(ctx *gin.Context, infoMap map[string]apiKeyInfo, apiKeyID string) apiKeyInfo {
	if info, ok := infoMap[apiKeyID]; ok {
		return info
	}
	deleted := gin_util.I18nKey(ctx, "app_statistic_api_key_deleted")
	return apiKeyInfo{
		name: deleted,
		key:  deleted,
	}
}
