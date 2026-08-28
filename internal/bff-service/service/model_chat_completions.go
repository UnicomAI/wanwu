package service

import (
	"fmt"
	"net/http"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func ModelChatCompletions(ctx *gin.Context, modelID string, req *mp_common.LLMReq, lineProcessor func(*mp_common.LLMResp) string) {
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

	// 校验model字段
	if req != nil {
		if req.Model != modelInfo.Model {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: model mismatch!", modelInfo.ModelId)))
			return
		}
	}

	// llm config
	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: %v", modelInfo.ModelId, err)))
		return
	}

	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: invalid provider", modelInfo.ModelId)))
		return
	}
	startTime := time.Now()
	requestBody := MarshalStatisticBody(req)
	// chat completions
	llmReq, err := iLLM.NewReq(req)
	if err != nil {
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, false, requestBody, err)
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}
	resp, sseCh, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReq)
	if err != nil {
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, llmReq.Stream(), requestBody, err)
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}
	// unary
	if !llmReq.Stream() {
		if data, ok := resp.ConvertResp(); ok {
			var retStr = resp.String()
			if lineProcessor != nil {
				retStr = lineProcessor(data)
			}
			status := http.StatusOK
			ctx.Set(gin_util.STATUS, status)
			ctx.Set(gin_util.RESULT, retStr)
			ctx.JSON(status, data)

			costs := int(time.Since(startTime).Milliseconds())
			finishReason := ""
			if len(data.Choices) > 0 {
				finishReason = data.Choices[0].FinishReason
			}
			go func() {
				defer util.PrintPanicStack()
				recordModelStatisticV2(detachedCtx, modelInfo,
					data.Usage.PromptTokens, data.Usage.CompletionTokens, data.Usage.TotalTokens, costs, 0, false,
					int64(status), requestBody, retStr, finishReason, "")
			}()
			return
		}
		// 非流式调用失败
		errMsg := fmt.Sprintf("model %v chat completions err: invalid resp", modelInfo.ModelId)
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, false, requestBody, fmt.Errorf("%s", errMsg))
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, errMsg))
		return
	}
	// stream
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	var answer string
	var finishReason string
	var (
		firstTokenTime    time.Time
		firstTokenLatency int
		promptTokens      int
		completionTokens  int
		totalTokens       int
	)
	var data *mp_common.LLMResp
	for sseResp := range sseCh {
		data, ok = sseResp.ConvertResp()
		dataStr := ""
		if ok && data != nil {
			if len(data.Choices) > 0 && data.Choices[0].FinishReason != "" {
				finishReason = data.Choices[0].FinishReason
			}
			if len(data.Choices) > 0 && data.Choices[0].Delta != nil {
				delta := data.Choices[0].Delta
				answer = answer + delta.Content
			}

			if lineProcessor != nil {
				dataStr = fmt.Sprintf("data: %v\n", lineProcessor(data))
			} else {
				dataStr = sseResp.String()
			}
			if firstTokenTime.IsZero() {
				firstTokenTime = time.Now()
				firstTokenLatency = int(time.Since(startTime).Milliseconds())
				ctx.Set(gin_util.FIRST_RESP_LATENCY, int64(firstTokenLatency))
			}
			promptTokens = data.Usage.PromptTokens
			completionTokens = data.Usage.CompletionTokens
			totalTokens = data.Usage.TotalTokens
		} else {
			dataStr = fmt.Sprintf("%v\n", sseResp.String())
		}
		// log.Debugf("model %v chat completions sse: %v", modelInfo.ModelId, dataStr)
		if _, err = ctx.Writer.Write([]byte(dataStr)); err != nil {
			log.Errorf("model %v chat completions sse err: %v", modelInfo.ModelId, err)
		}
		ctx.Writer.Flush()
	}
	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, answer)
	go func() {
		defer util.PrintPanicStack()
		recordModelStatisticV2(detachedCtx, modelInfo,
			promptTokens, completionTokens, totalTokens, 0, firstTokenLatency, true,
			http.StatusOK, requestBody, answer, finishReason, "")
	}()
}
