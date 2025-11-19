package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/gin-gonic/gin"
)

func ModelChatCompletions(ctx *gin.Context, modelID string, req *mp_common.LLMReq) {
	// modelInfo by modelID
	modelInfo, err := model.GetModelById(ctx.Request.Context(), &model_service.GetModelByIdReq{ModelId: modelID})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// Verify model fields
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

	// chat completions
	llmReq, err := iLLM.NewReq(req)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions NewReq err: %v", modelInfo.ModelId, err)))
		return
	}
	resp, sseCh, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReq)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: %v", modelInfo.ModelId, err)))
		return
	}
	// unary
	if !llmReq.Stream() {
		if data, ok := resp.ConvertResp(); ok {
			status := http.StatusOK
			ctx.Set(gin_util.STATUS, status)
			ctx.Set(gin_util.RESULT, resp.String())
			ctx.JSON(status, data)
			return
		}
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: invalid resp", modelInfo.ModelId)))
		return
	}
	// stream
	var answer string
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	var (
		firstFlag = false // The starting identifier of the thinking chain. The default thinking chain has not started.
		endFlag   = false // The end identifier of the thinking chain. The default thinking chain is not ended.
	)
	var data *mp_common.LLMResp
	for sseResp := range sseCh {
		data, ok = sseResp.ConvertResp()
		dataStr := ""
		if ok && data != nil {
			if len(data.Choices) > 0 && data.Choices[0].Delta != nil {
				answer = answer + data.Choices[0].Delta.Content
				delta := data.Choices[0].Delta
				if firstFlag && !endFlag && delta.ReasoningContent != nil {
					delta.Content = delta.Content + *delta.ReasoningContent
				}
				if !endFlag && delta.Content != "" && ((delta.ReasoningContent != nil &&
					*delta.ReasoningContent == "") || delta.ReasoningContent == nil) && firstFlag {
					delta.Content = "\n</think>\n" + delta.Content
					endFlag = true
				}
				if !firstFlag && delta.ReasoningContent != nil && *delta.ReasoningContent != "" && delta.Content == "" {
					delta.Content = "<think>\n" +
						delta.Content + *delta.ReasoningContent
					firstFlag = true
				}
			}
			dataByte, _ := json.Marshal(data)
			dataStr = fmt.Sprintf("data: %v\n", string(dataByte))
		} else {
			dataStr = fmt.Sprintf("%v\n", sseResp.String())
		}
		//log.Infof("model %v chat completions sse: %v", modelInfo.ModelId, dataStr)
		if _, err = ctx.Writer.Write([]byte(dataStr)); err != nil {
			log.Errorf("model %v chat completions sse err: %v", modelInfo.ModelId, err)
		}
		ctx.Writer.Flush()
	}
	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, answer)
}
