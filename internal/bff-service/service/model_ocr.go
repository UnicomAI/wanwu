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

func ModelOcr(ctx *gin.Context, modelID string, req *mp_common.OcrReq) {
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
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

	req.Model = &modelInfo.Model
	// ocr config
	ocr, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v ocr err: %v", modelInfo.ModelId, err)))
		return
	}
	iOcr, ok := ocr.(mp.IOcr)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v ocr err: invalid provider", modelInfo.ModelId)))
		return
	}

	ocrReq, err := iOcr.NewReq(req)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v ocr NewReq err: %v", modelInfo.ModelId, err)))
		return
	}
	startTime := time.Now()
	resp, err := iOcr.Ocr(ctx, ocrReq)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v ocr err: %v", modelInfo.ModelId, err)))
		return
	}
	if data, ok := resp.ConvertResp(); ok {
		status := http.StatusOK
		ctx.Set(gin_util.STATUS, status)
		//ctx.Set(config.RESULT, resp.String())
		ctx.JSON(status, data)
		costs := int(time.Since(startTime).Milliseconds())
		go func() {
			defer util.PrintPanicStack()
			recordModelStatistic(detachedCtx, modelInfo, true, 0, 0, 0, costs, 0, false)
		}()
		return
	}
	go func() {
		defer util.PrintPanicStack()
		recordModelStatistic(detachedCtx, modelInfo, false, 0, 0, 0, 0, 0, false)
	}()
	gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v ocr err: invalid resp", modelInfo.ModelId)))
}
