package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
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
		if !(category == "" || category == "all") && !strings.Contains(promptCfg.Category, category) {
			continue
		}
		promptTemplateList = append(promptTemplateList, buildPromptTempDetail(*promptCfg))
	}
	fmt.Println()
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
	// 获取模型信息 [EN] Get model information
	modelInfo, err := model.GetModelById(ctx.Request.Context(), &model_service.GetModelByIdReq{ModelId: req.ModelId})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	// 构建请求信息 [EN] Build request information
	var stream bool = true
	reqInfo := &mp_common.LLMReq{
		Model: modelInfo.Model,
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

	// 配置模型参数 [EN] Configure model parameters
	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return
	}
	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: invalid provider", modelInfo.ModelId)))
		return
	}

	// chat completions
	llmReq, err := iLLM.NewReq(reqInfo)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions NewReq err: %v", modelInfo.ModelId, err)))
		return
	}
	_, sseCh, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReq)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v chat completions err: %v", modelInfo.ModelId, err)))
		return
	}

	// stream
	var answer string
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	var data *mp_common.LLMResp
	var inThink bool = false // 是否在思考标签内 [EN] Are you thinking about tags?

	for sseResp := range sseCh {
		data, ok = sseResp.ConvertResp()
		var dataStr string
		var shouldSend bool = true // 标记是否应该发送此响应 [EN] Flag whether this response should be sent

		if ok && data != nil {
			currentResponse := "" // 记录当前流式增量内容 [EN] Record the current streaming delta content
			if len(data.Choices) > 0 && data.Choices[0].Delta != nil {
				content := data.Choices[0].Delta.Content

				// 过滤思考过程 [EN] filter thought process
				if inThink {
					// 当前在思考标签内，检查是否遇到结束标签 [EN] Currently inside the think tag, check if an end tag is encountered
					if strings.Contains(content, "</think>") {
						inThink = false
						parts := strings.SplitN(content, "</think>", 2)
						// 找到</think>之后的内容 [EN] Find the content after </think>
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
					// 不在思考标签内，检查是否遇到开始标签 [EN] Not within the think tag, check if a start tag is encountered
					if strings.Contains(content, "<think>") {
						// 检查是否有<think>之前的内容 [EN] Check if there is content before <think>
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
								// </think>之后还有内容，需要返回 [EN] </think>There is still content after that and needs to be returned
								answer = answer + endParts[1]
								currentResponse = currentResponse + endParts[1]
							} else if currentResponse == "" {
								shouldSend = false
							}
							inThink = false
						}
					} else {
						// 没有思考标签，直接返回内容 [EN] No thinking about tags, return content directly
						answer = answer + content
						currentResponse = content
					}
				}
			}

			// 发送响应 [EN] Send response
			if shouldSend {
				// 构建目标结构 [EN] Build target structure
				streamData := response.CustomPromptOpt{
					Code:     data.Code,
					Message:  "success",
					Response: currentResponse,
					Finish:   0,
					Usage:    &data.Usage,
				}
				if len(data.Choices) > 0 {
					if data.Choices[0].FinishReason == "" {
						streamData.Finish = 0 // 继续生成 [EN] Continue to generate
					} else if data.Choices[0].FinishReason == "stop" {
						streamData.Finish = 1 // 结束标志 [EN] end sign
					}
				}

				dataByte, _ := json.Marshal(streamData)
				dataStr = fmt.Sprintf("data: %v\n", string(dataByte))
			}
		} else {
			dataStr = fmt.Sprintf("%v\n", sseResp.String())
		}

		// 写入 [EN] write
		if dataStr != "" {
			if _, err = ctx.Writer.Write([]byte(dataStr)); err != nil {
				log.Errorf("model %v chat completions sse err: %v", modelInfo.ModelId, err)
			}
			ctx.Writer.Flush()
		}
	}

	if len(answer) == 0 {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "answer is empty"))
		return
	}

	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, answer)
}

// --- internal ---
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
