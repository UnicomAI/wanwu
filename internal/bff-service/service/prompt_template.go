package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func CreatePromptByTemplate(ctx *gin.Context, userID, orgID string, req request.CreatePromptByTemplateReq) (*response.PromptIDData, error) {
	promptCfg, exist := config.Cfg().PromptTemp(req.TemplateId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_prompt_template_detail", "get prompt template detail empty")
	}
	promptIDResp, err := assistant.CustomPromptCreate(ctx.Request.Context(), &assistant_service.CustomPromptCreateReq{
		AvatarPath: req.Avatar.Key,
		Name:       req.Name,
		Desc:       req.Desc,
		Prompt:     promptCfg.Prompt,
		Identity: &assistant_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.PromptIDData{
		PromptId: promptIDResp.CustomPromptId,
	}, nil
}

func GetPromptTemplateList(ctx *gin.Context, category, name string) (*response.ListResult, error) {
	var promptTemplateList []*response.PromptTemplateDetail
	for _, promptCfg := range config.Cfg().PromptTemplates {
		if name != "" && !strings.Contains(promptCfg.Name, name) {
			continue
		}
		if category != "" && category != "all" && !strings.Contains(promptCfg.Category, category) {
			continue
		}
		promptTemplateList = append(promptTemplateList, buildPromptTempDetail(*promptCfg))
	}
	return &response.ListResult{
		List:  promptTemplateList,
		Total: int64(len(promptTemplateList)),
	}, nil
}

func GetPromptTemplateDetail(ctx *gin.Context, templateId string) (*response.PromptTemplateDetail, error) {
	promptCfg, exist := config.Cfg().PromptTemp(templateId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_prompt_template_detail", "get prompt template detail empty")
	}
	return buildPromptTempDetail(promptCfg), nil
}

func GetPromptOptimize(ctx *gin.Context, userID, orgID string, req request.PromptOptimizeReq) {
	// 构建请求信息
	var stream = true
	reqInfo := &mp_common.LLMReq{
		Messages: []mp_common.OpenAIReqMsg{
			{
				Role:    mp_common.MsgRoleSystem,
				Content: strings.ReplaceAll(config.Cfg().PromptEngineering.Optimization, "{{message}}", req.Prompt),
			},
			{
				Role:    mp_common.MsgRoleUser,
				Content: req.Prompt,
			},
		},
		Stream: &stream,
	}
	getPromptCustom(ctx, userID, orgID, req.ModelId, reqInfo)
}

func GetPromptReason(ctx *gin.Context, userID, orgID string, req request.PromptReasonReq) {
	// 构建提示词推理请求信息
	var stream = true
	reqInfo := &mp_common.LLMReq{
		Messages: []mp_common.OpenAIReqMsg{
			{
				Role:    mp_common.MsgRoleUser,
				Content: req.Prompt,
			},
		},
		Stream: &stream,
	}
	getPromptCustom(ctx, userID, orgID, req.ModelId, reqInfo)
}

func GetPromptEvaluate(ctx *gin.Context, userID, orgID string, req request.PromptEvaluateReq) {
	// 构建提示词推理请求信息
	var stream = true
	content := strings.ReplaceAll(config.Cfg().PromptEngineering.Evaluation, "{{target}}", req.ExpectedOutput)
	content = strings.ReplaceAll(content, "{{answer}}", req.Answer)

	// 构建提示词评估请求信息
	evaReqInfo := &mp_common.LLMReq{
		Messages: []mp_common.OpenAIReqMsg{
			{
				Role:    mp_common.MsgRoleSystem,
				Content: content,
			},
			{
				Role:    mp_common.MsgRoleUser,
				Content: "目标回答： " + req.ExpectedOutput + "\n 待评估回答：" + req.Answer,
			},
		},
		Stream: &stream,
	}
	getPromptCustom(ctx, userID, orgID, req.ModelId, evaReqInfo)
}

// --- internal ---
// getPromptCustom 提示词优化/推理/评估共用流式调用；写模型统计 + 应用统计（prompt 板块级，appId 为空）。
func getPromptCustom(ctx *gin.Context, userID, orgID, modelId string, reqInfo *mp_common.LLMReq) {
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	modelRequestBody := MarshalStatisticBody(reqInfo)
	var (
		statErr           error
		firstTokenLatency int
		startTime         time.Time
	)
	defer func() {
		statusCode, failureReason := appStreamStatisticStatus(statErr, "")
		source := resolveAppStatisticSource(detachedCtx, constant.BizSourceWeb)
		// 流式：记录真实首 token 时延；未打到则保持 0（不用总耗时冒充 TTFT）
		go func() {
			defer util.PrintPanicStack()
			RecordAppStatistic(detachedCtx, userID, orgID, "", "", constant.BizModuleResourcePrompt,
				statusCode, failureReason, true, int64(firstTokenLatency), 0, source, modelRequestBody, "", "", "")
		}()
	}()

	// 获取模型信息
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelId})
	if err != nil {
		statErr = err
		gin_util.Response(ctx, nil, err)
		return
	}
	if !modelInfo.IsActive {
		statErr = grpc_util.ErrorStatus(errs.Code_BFFModelStatus, modelInfo.ModelId)
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(errs.Code_BFFModelStatus, modelInfo.ModelId))
		return
	}
	reqInfo.Model = modelInfo.Model

	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		statErr = err
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, false, modelRequestBody, err)
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error()))
		return
	}

	// 判断模型是否支持深度思考，若支持则默认添加 enable_thinking: false
	if llm != nil {
		llmValue := reflect.ValueOf(llm).Elem()
		thinkingField := llmValue.FieldByName("ThinkingSupport")
		if thinkingField.IsValid() && thinkingField.Kind() == reflect.String {
			thinkingSupport := thinkingField.String()
			if thinkingSupport == "support" {
				enableThinking := false
				reqInfo.EnableThinking = &enableThinking
			}
		}
	}

	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		statErr = fmt.Errorf("model %v chat completions err: invalid provider", modelInfo.ModelId)
		errMsg := fmt.Sprintf("model %v chat completions err: invalid provider", modelInfo.ModelId)
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, false, modelRequestBody, fmt.Errorf("%s", errMsg))
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, errMsg))
		return
	}

	// chat completions
	llmReq, err := iLLM.NewReq(reqInfo)
	if err != nil {
		statErr = err
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, false, modelRequestBody, err)
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error()))
		return
	}
	startTime = time.Now()
	_, sseCh, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReq)
	if err != nil {
		statErr = err
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, false, modelRequestBody, err)
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error()))
		return
	}

	// stream
	var answer string
	var finishReason string

	var (
		firstTokenTime   time.Time
		promptTokens     int
		completionTokens int
		totalTokens      int
	)

	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	var data *mp_common.LLMResp
	var inThink = false // 是否在思考标签内

	for sseResp := range sseCh {
		data, ok = sseResp.ConvertResp()
		var dataStr string
		var shouldSend = true // 标记是否应该发送此响应

		if ok && data != nil {
			if len(data.Choices) > 0 && data.Choices[0].FinishReason != "" {
				finishReason = data.Choices[0].FinishReason
			}
			currentResponse := "" // 记录当前流式增量内容
			if len(data.Choices) > 0 && data.Choices[0].Delta != nil {
				content := data.Choices[0].Delta.Content

				// 过滤思考过程
				if inThink {
					// 当前在思考标签内，检查是否遇到结束标签
					if strings.Contains(content, "</think>") {
						inThink = false
						parts := strings.SplitN(content, "</think>", 2)
						// 找到</think>之后的内容
						if len(parts) > 1 && parts[1] != "" {
							filteredContent := parts[1]
							answer = answer + filteredContent
							currentResponse = filteredContent
						} else {
							shouldSend = false
						}
					} else {
						shouldSend = false
					}
				} else {
					// 不在思考标签内，检查是否遇到开始标签
					if strings.Contains(content, "<think>") {
						// 检查是否有<think>之前的内容
						parts := strings.SplitN(content, "<think>", 2)
						if len(parts) > 0 && parts[0] != "" {
							filteredContent := parts[0]
							answer = answer + filteredContent
							currentResponse = filteredContent
						} else {
							shouldSend = false
						}
						inThink = true

						if len(parts) > 1 && strings.Contains(parts[1], "</think>") {
							endParts := strings.SplitN(parts[1], "</think>", 2)
							if len(endParts) > 1 && endParts[1] != "" {
								// </think>之后还有内容，需要返回
								answer = answer + endParts[1]
								currentResponse = currentResponse + endParts[1]
							} else if currentResponse == "" {
								shouldSend = false
							}
							inThink = false
						}
					} else {
						// 没有思考标签，直接返回内容
						answer = answer + content
						currentResponse = content
					}
				}
			}

			// 发送响应
			if shouldSend {
				// 构建目标结构
				streamData := response.CustomPromptOpt{
					Code:     data.Code,
					Message:  "success",
					Response: currentResponse,
					Finish:   0,
					Usage:    &data.Usage,
				}
				if len(data.Choices) > 0 {
					switch data.Choices[0].FinishReason {
					case "":
						streamData.Finish = 0 // 继续生成
					case "stop":
						streamData.Finish = 1 // 结束标志
					}
				}

				dataByte, _ := json.Marshal(streamData)
				dataStr = fmt.Sprintf("data: %v\n", string(dataByte))
			}
			if firstTokenTime.IsZero() {
				firstTokenTime = time.Now()
				firstTokenLatency = int(time.Since(startTime).Milliseconds())
			}
			promptTokens = data.Usage.PromptTokens
			completionTokens = data.Usage.CompletionTokens
			totalTokens = data.Usage.TotalTokens
		} else {
			// 流式过程中，大模型sse返回的这一行是空行，即sseResp.String()==""；前端正常展示，也需要这个空行
			dataStr = fmt.Sprintf("%v\n", sseResp.String())
		}

		// 写入
		if dataStr != "" {
			if _, err = ctx.Writer.Write([]byte(dataStr)); err != nil {
				log.Errorf("model %v chat completions sse err: %v", modelInfo.ModelId, err)
			}
			ctx.Writer.Flush()
		}
	}

	if len(answer) == 0 {
		statErr = fmt.Errorf("answer is empty")
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, true, modelRequestBody, fmt.Errorf("answer is empty"))
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "answer is empty"))
		return
	}

	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, answer)
	go func() {
		defer util.PrintPanicStack()
		recordModelStatisticV2(detachedCtx, modelInfo,
			promptTokens, completionTokens, totalTokens, 0, firstTokenLatency, true,
			http.StatusOK, modelRequestBody, answer, finishReason, "")
	}()
}

func buildPromptTempDetail(wtfCfg config.PromptTempConfig) *response.PromptTemplateDetail {
	iconUrl := config.Cfg().DefaultIcon.PromptIcon
	return &response.PromptTemplateDetail{
		TemplateId: wtfCfg.TemplateId,
		Category:   wtfCfg.Category,
		Author:     wtfCfg.Author,
		Prompt:     wtfCfg.Prompt,
		AppBriefConfig: request.AppBriefConfig{
			Avatar: request.Avatar{Path: iconUrl},
			Name:   wtfCfg.Name,
			Desc:   wtfCfg.Desc,
		},
	}
}
