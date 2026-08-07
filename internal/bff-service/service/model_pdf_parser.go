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

func ModelPdfParser(ctx *gin.Context, modelID string, req *mp_common.PdfParserReq) {
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

	// pdfParser config
	pdfParser, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v pdfParser err: %v", modelInfo.ModelId, err)))
		return
	}
	iPdfParser, ok := pdfParser.(mp.IPdfParser)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v pdfParser err: invalid provider", modelInfo.ModelId)))
		return
	}

	pdfParserReq, err := iPdfParser.NewReq(req)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v pdfParser NewReq err: %v", modelInfo.ModelId, err)))
		return
	}
	requestBody := MarshalStatisticBody(req)
	startTime := time.Now()
	resp, err := iPdfParser.PdfParser(ctx, pdfParserReq)
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
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2(detachedCtx, modelInfo, 0, 0, 0, costs, 0, false,
				http.StatusOK, requestBody, responseBody, "", "")
		}()
		return
	}
	errMsg := fmt.Sprintf("model %v pdfParser err: invalid resp", modelInfo.ModelId)
	go func() {
		defer util.PrintPanicStack()
		recordModelStatisticV2Failure(detachedCtx, modelInfo, false, requestBody, fmt.Errorf("%s", errMsg))
	}()
	gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, errMsg))
}
