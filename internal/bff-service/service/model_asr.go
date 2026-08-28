package service

import (
	"fmt"
	"net/http"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func ModelSyncAsr(ctx *gin.Context, modelID string, req *mp_common.SyncAsrReq) {
	// modelInfo by modelID
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelID})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if !modelInfo.IsActive {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFModelStatus, modelInfo.ModelId))
		return
	}

	// 校验model字段
	if req != nil {
		if req.Model != modelInfo.Model {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v sync_asr err: model mismatch!", modelInfo.ModelId)))
			return
		}
	}
	modelSyncAsr(ctx, modelInfo, modelID, modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig, req)
}

func modelSyncAsr(ctx *gin.Context, modelInfo *model_service.ModelInfo, modelId, provider, modelType, providerConfig string, req *mp_common.SyncAsrReq) {
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	// sync_asr config
	sync_asr, err := mp.ToModelConfig(provider, modelType, providerConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v sync_asr err: %v", modelId, err)))
		return
	}
	iSyncAsr, ok := sync_asr.(mp.ISyncAsr)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v sync_asr err: invalid provider", modelId)))
		return
	}
	requestBody := MarshalStatisticBody(req)
	startTime := time.Now()
	asrReq, err := iSyncAsr.NewReq(req)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v sync_asr NewReq err: %v", modelId, err)))
		return
	}
	resp, err := iSyncAsr.SyncAsr(ctx, asrReq)
	if err != nil {
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, false, requestBody, err)
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}
	if data, ok := resp.ConvertResp(); ok {
		status := http.StatusOK
		ctx.Set(gin_util.STATUS, status)
		//ctx.Set(config.RESULT, resp.String())
		ctx.JSON(status, data)
		costs := int(time.Since(startTime).Milliseconds())
		responseBody := MarshalStatisticBody(data)
		finishReason := ""
		if len(data.Choices) > 0 {
			finishReason = data.Choices[0].FinishReason
		}
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2(detachedCtx, modelInfo, 0, 0, 0, costs, 0, false,
				http.StatusOK, requestBody, responseBody, finishReason, "")
		}()
		return
	}
	errMsg := fmt.Sprintf("model %v sync_asr err: invalid resp", modelId)
	go func() {
		defer util.PrintPanicStack()
		recordModelStatisticV2Failure(detachedCtx, modelInfo, false, requestBody, fmt.Errorf("%s", errMsg))
	}()
	gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, errMsg))
}
