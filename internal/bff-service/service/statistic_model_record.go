package service

import (
	"context"
	"encoding/json"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/UnicomAI/wanwu/pkg/statistic"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
)

// recordModelStatisticV2 记录模型调用统计（V2：明细 + Redis 聚合）。
// statusCode/failureReason 约定与应用统计一致：成功传 200；失败优先
// GrpcErrorToHTTPStatus(err)（见 recordModelStatisticV2Failure）。isSuccess 由 statusCode 反推，
// token 按模型类型归一化（OCR/PDF/ASR 清零）。
func recordModelStatisticV2(ctx context.Context, modelInfo *model_service.ModelInfo,
	promptTokens, completionTokens, totalTokens, costs, firstTokenLatency int,
	isStream bool, statusCode int64, requestBody, responseBody, finishReason, failureReason string) {
	promptTokens, completionTokens, totalTokens = normalizeModelStatisticTokens(modelInfo.GetModelType(), promptTokens, completionTokens, totalTokens)
	isSuccess := statistic.IsSuccess(statusCode)
	statCtx, err := trace_util.ParseStatisticContext(ctx)
	if err != nil {
		log.Errorf("record model statistic v2 parse context err: %v", err)
		return
	}
	_, err = app.RecordModelStatisticV2(ctx, &app_service.RecordModelStatisticV2Req{
		TraceId:             statCtx.TraceID,
		UserId:              statCtx.UserID,
		OrgId:               statCtx.OrgID,
		Source:              statCtx.Source,
		Module:              statCtx.Module,
		AppId:               statCtx.AppID,
		AppType:             statCtx.AppType,
		ApiKey:              statCtx.APIKey,
		ApiKeyId:            statCtx.APIKeyID,
		MethodPath:          statCtx.MethodPath,
		ModuleCreatorUserId: statCtx.ModuleCreatorUserID,
		ModuleCreatorOrgId:  statCtx.ModuleCreatorOrgID,
		ModelId:             modelInfo.ModelId,
		Model:               modelInfo.Model,
		Provider:            modelInfo.Provider,
		ModelType:           modelInfo.ModelType,
		ModelCreatorUserId:  modelInfo.UserId,
		ModelCreatorOrgId:   modelInfo.OrgId,
		PromptTokens:        int64(promptTokens),
		CompletionTokens:    int64(completionTokens),
		TotalTokens:         int64(totalTokens),
		FirstTokenLatency:   int64(firstTokenLatency),
		Costs:               int64(costs),
		IsSuccess:           isSuccess,
		IsStream:            isStream,
		StatusCode:          statusCode,
		RequestBody:         maybeRecordBody(requestBody),
		ResponseBody:        maybeRecordBody(responseBody),
		FinishReason:        finishReason,
		FailureReason:       failureReason,
	})
	if err != nil {
		log.Errorf("record model statistic v2 modelId %v err: %v", modelInfo.ModelId, err)
	}
	// 模型体验等 module=model 路径：同步写入应用维度板块级行（appId 为空），与 prompt/wga 一致。
	if statCtx.Module == constant.BizModuleModel {
		streamCosts, nonStreamCosts := int64(0), int64(0)
		if isStream {
			streamCosts = int64(firstTokenLatency)
		} else {
			nonStreamCosts = int64(costs)
		}
		RecordAppStatistic(ctx, statCtx.UserID, statCtx.OrgID, "", "", constant.BizModuleModel,
			statusCode, failureReason, isStream, streamCosts, nonStreamCosts, statCtx.Source, requestBody, responseBody, "", "")
	}
}

func maybeRecordBody(body string) string {
	if config.Cfg().Statistic.RecordBody == 0 {
		return ""
	}
	return body
}

// recordModelStatisticV2Failure 记录失败调用，statusCode 映射与应用统计一致（GrpcErrorToHTTPStatus）。
// 调用点均传非 nil err；GrpcErrorToHTTPStatus 对失败 err 返回对应状态码与消息。
func recordModelStatisticV2Failure(ctx context.Context, modelInfo *model_service.ModelInfo,
	isStream bool, requestBody string, err error) {
	statusCode, failureReason := GrpcErrorToHTTPStatus(err)
	recordModelStatisticV2(ctx, modelInfo, 0, 0, 0, 0, 0, isStream, statusCode, requestBody, "", "", failureReason)
}

func normalizeModelStatisticTokens(modelType string, promptTokens, completionTokens, totalTokens int) (int, int, int) {
	switch modelType {
	case mp.ModelTypeOcr, mp.ModelTypePdfParser, mp.ModelTypeSyncAsr:
		return 0, 0, 0
	default:
		return promptTokens, completionTokens, totalTokens
	}
}

// MarshalStatisticBody 将整体请求/响应 JSON 序列化，供模型/应用统计明细落库。
func MarshalStatisticBody(v any) string {
	if v == nil {
		return ""
	}
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}
