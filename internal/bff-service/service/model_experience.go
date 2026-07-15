package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/ahocorasick"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	sse_connector "github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector"
	sse_model "github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/model"
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/store"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func ModelExperienceLLM(ctx *gin.Context, userId, orgId, clientId string, req *request.ModelExperienceLlmRequest) {

	// 敏感词检测 - 输入检测
	matchDicts, err := BuildSensitiveDict(ctx, nil, false)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	matchResults, err := ahocorasick.ContentMatch(req.Content, matchDicts, true)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFSensitiveWordCheck, err.Error()))
		return
	}
	if len(matchResults) > 0 {
		if matchResults[0].Reply != "" {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req", matchResults[0].Reply))
			return
		}
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req_default_reply"))
		return
	}
	// save query
	if _, err := model.SaveModelExperienceDialogRecord(ctx.Request.Context(), &model_service.SaveModelExperienceDialogRecordReq{
		UserId:            userId,
		OrgId:             orgId,
		ModelExperienceId: req.ModelExperienceId,
		ModelId:           req.ModelId,
		SessionId:         req.SessionId,
		OriginalContent:   req.Content,
		Role:              string(mp_common.MsgRoleUser),
		FileInfo:          toJsonFromModelExperienceFileInfo(req.FileInfo),
	}); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	// model info
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: req.ModelId})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if !modelInfo.IsActive {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFModelStatus, modelInfo.ModelId))
		return
	}

	// 检查模型是否支持图文问答
	hasVisionSupport, err := checkModelExperienceModelVisionSupport(modelInfo)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	// 校验文件：只有支持图文问答的模型才能上传文件，且需要校验文件大小和类型
	if len(req.FileInfo) > 0 {
		if !hasVisionSupport {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "model does not support vision"))
			return
		}
		if err := validateModelExperienceImageFiles(ctx, modelInfo, req.FileInfo); err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
	}

	// dialog records
	recordsResp, err := model.GetModelExperienceDialogRecords(ctx.Request.Context(), &model_service.GetModelExperienceDialogRecordsReq{
		UserId: userId,
		OrgId:  orgId,
		// 常规模型对话记录（非模型对比时），modelExperienceId与sessionId非空
		// 模型对比时临时存储对话记录，modelExperienceId前端传空，sessionId非空
		ModelExperienceId: req.ModelExperienceId,
		SessionId:         req.SessionId,
	})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	var messages []mp_common.OpenAIReqMsg
	for _, record := range recordsResp.Records {
		content := record.HandledContent
		if content == "" {
			content = record.OriginalContent
		}
		// 解析历史记录中的 fileInfo
		var fileInfo []request.ConversionStreamFile
		// 只有支持图文问答的模型才处理文件
		if hasVisionSupport && record.FileInfo != "" {
			fileInfo = replaceModelExperienceMinioUrl(record.FileInfo)
		}
		messages = append(messages, mp_common.OpenAIReqMsg{
			Role:    mp_common.MsgRole(record.Role),
			Content: buildModelExperienceMultimodalContent(ctx, content, fileInfo),
		})
	}
	// 添加当前用户消息（支持多模态）
	var currentUserFileInfo []request.ConversionStreamFile
	if hasVisionSupport {
		currentUserFileInfo = req.FileInfo
	}
	userMsg := mp_common.OpenAIReqMsg{
		Role:    mp_common.MsgRoleUser,
		Content: buildModelExperienceMultimodalContent(ctx, req.Content, currentUserFileInfo),
	}

	messages = append(messages, userMsg)

	// 构造LLM请求
	stream := true
	llmReq := &mp_common.LLMReq{
		Model:    modelInfo.Model,
		Messages: messages,
		Stream:   &stream,
	}
	if req.TemperatureEnable {
		llmReq.Temperature = &req.Temperature
	}
	if req.TopPEnable {
		llmReq.TopP = &req.TopP
	}
	if req.PresencePenaltyEnable {
		llmReq.PresencePenalty = &req.PresencePenalty
	}
	if req.FrequencyPenaltyEnable {
		llmReq.FrequencyPenalty = &req.FrequencyPenalty
	}
	if req.MaxTokensEnable {
		maxTokens := int(req.MaxTokens)
		llmReq.MaxTokens = &maxTokens
	}
	llmReq.EnableThinking = req.ThinkingEnable

	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}
	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}
	startTime := time.Now()

	// chat completions
	iLLMReq, err := iLLM.NewReq(llmReq)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}

	// 创建 SSE Session 用于断点续传（仅当 clientId 存在时）
	session := &sse_model.Session{ConversationID: req.SessionId, ClientID: clientId}
	sseSessionManager := sse_connector.NewSSESession(ctx.Request.Context(), session, store.NewMemoryStore())
	bgCtx := sseSessionManager.GetBgContext()

	_, sseCh, err := iLLM.ChatCompletions(bgCtx, iLLMReq)
	if err != nil {
		go func() {
			defer util.PrintPanicStack()
			recordModelStatistic(bgCtx, modelInfo, false, 0, 0, 0, 0, 0, false)
			_ = sseSessionManager.Cancel()
		}()
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error()))
		return
	}
	// stream
	var answer string
	var reasonContent string
	var (
		firstTokenLatency int
		promptTokens      int
		completionTokens  int
		totalTokens       int
	)
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")

	// 消费 sseCh 时一边统计 token，一边转成 rawCh 交给敏感词处理。
	rawCh := make(chan string, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(rawCh)
		firstTokenReceived := false
		for sseResp := range sseCh {
			resp, ok := sseResp.ConvertResp()
			if ok && resp != nil {
				if !firstTokenReceived {
					firstTokenReceived = true
					firstTokenLatency = int(time.Since(startTime).Milliseconds())
				}
				promptTokens = resp.Usage.PromptTokens
				completionTokens = resp.Usage.CompletionTokens
				totalTokens = resp.Usage.TotalTokens
			}
			// ConvertResp 失败的空行也是 SSE 事件分隔符，必须保留。
			dataStr := sseResp.String()
			if !ok {
				dataStr = fmt.Sprintf("%v\n", dataStr)
			}
			// 敏感词过滤结束后继续 drain sseCh 统计 token，但不再阻塞写 rawCh。
			select {
			case rawCh <- dataStr:
			default:
			}
		}
		recordModelStatistic(bgCtx, modelInfo, true,
			promptTokens, completionTokens, totalTokens, 0, firstTokenLatency, true)
	}()

	// 敏感词过滤（必须过滤，全局敏感词）
	outputCh := ProcessSensitiveWordsWithCallback(ctx, rawCh, matchDicts, &modelExperienceSensitiveService{}, func(messageId string, sensitiveMsg string) {
		//触发sse cancel
		_ = sseSessionManager.Cancel()
	})
	// 从过滤后的 channel 写入 SSE 并累加 answer
	for dataStr := range outputCh {

		_ = sseSessionManager.Publish(&sse_model.Message{Data: dataStr}, modelExperienceSSECompactProcessor())

		if _, err = ctx.Writer.Write([]byte(dataStr)); err != nil {
			log.Errorf("model experience write sse err: %v", err)
		}
		ctx.Writer.Flush()

		// 解析 SSE 内容累加 answer
		if strings.HasPrefix(dataStr, "data:") {
			content := strings.TrimSpace(strings.TrimPrefix(dataStr, "data:"))
			if content == "[DONE]" {
				continue
			}
			var resp mp_common.LLMResp
			if err := json.Unmarshal([]byte(content), &resp); err == nil {
				if len(resp.Choices) > 0 && resp.Choices[0].Delta != nil {
					delta := resp.Choices[0].Delta
					answer = answer + delta.Content
					if delta.ReasoningContent != nil {
						reasonContent = reasonContent + *delta.ReasoningContent
					}
				}
			}
		}
	}

	defer func() {
		// 如果没有累积任何内容，不保存
		if answer == "" && reasonContent == "" {
			return
		}
		if _, err := model.SaveModelExperienceDialogRecord(trace_util.DetachContext(bgCtx), &model_service.SaveModelExperienceDialogRecordReq{
			UserId:            userId,
			OrgId:             orgId,
			ModelExperienceId: req.ModelExperienceId,
			ModelId:           req.ModelId,
			SessionId:         req.SessionId,
			OriginalContent:   answer,
			ReasoningContent:  reasonContent,
			Role:              string(mp_common.MsgRoleAssistant),
		}); err != nil {
			log.Errorf("model experience save record err: %v", err)
		}

		if err1 := sse_connector.Close(session); err1 != nil {
			log.Errorf("[ModelExperience] %v session %v user %v org %v session finish err: %v", req.ModelExperienceId, req.SessionId, userId, orgId, err1)
		}

	}()

	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, answer)

}

func ModelExperienceLLMConnect(ctx *gin.Context, userId, orgId, clientId string, req request.ModelExperienceLlmConnectRequest) error {
	chatCh, err := sse_connector.Connect[string](ctx, &sse_model.Session{ConversationID: req.SessionId, ClientID: clientId}, func(data *sse_model.Message) string {
		return data.Data.(string)
	})
	if err != nil {
		log.Infof("ModelExperienceLLMConnect session not found, sessionId: %s, treat as empty: %v", req.SessionId, err)
		return err
	}
	// 流式返回结果，模型体验的 SSE 数据已经是 data: {...}\n\n 格式，直接透传即可
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[ModelExperience] session %v user %v org %v recv", req.SessionId, userId, orgId), sse_util.DONE_MSG).
		WriteStream(chatCh, nil, buildModelExperienceRespLineProcessor(), nil)
	return nil
}

// buildModelExperienceRespLineProcessor 构造模型体验对话结果行处理器
// 模型体验的 SSE 数据已经是 data: {...}\n\n 格式，直接透传即可
func buildModelExperienceRespLineProcessor() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if strings.HasPrefix(lineText, "error:") {
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		return "data:" + lineText + "\n\n", false, nil
	}
}

func ModelExperienceLLMCancel(req request.ModelExperienceLlmCancelRequest, clientId string) error {
	return sse_connector.Close(&sse_model.Session{ConversationID: req.SessionId, ClientID: clientId})
}

func SaveModelExperienceDialog(ctx *gin.Context, userId, orgId string, req *request.ModelExperienceDialogRequest) (*response.ModelExperienceDialog, error) {
	// 将interface{}类型的ModelSetting转换为 json string
	var modelSettingStr string
	if req.ModelSetting != nil {
		jsonBytes, err := json.Marshal(req.ModelSetting)
		if err != nil {
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("Model settings serialization error: err: %v", err))
		}
		modelSettingStr = string(jsonBytes)
	}
	dialog, err := model.SaveModelExperienceDialog(ctx.Request.Context(), &model_service.SaveModelExperienceDialogReq{
		UserId:       userId,
		OrgId:        orgId,
		ModelId:      req.ModelId,
		SessionId:    req.SessionId,
		ModelSetting: modelSettingStr,
		Title:        req.Title,
	})
	if err != nil {
		return nil, err
	}
	return toModelExperienceDialog(dialog), nil
}

func ListModelExperienceDialogs(ctx *gin.Context, userId, orgId string) (*response.ListResult, error) {
	resp, err := model.GetModelExperienceDialogs(ctx.Request.Context(), &model_service.ListModelExperienceDialogReq{
		UserId: userId,
		OrgId:  orgId,
	})
	if err != nil {
		return nil, err
	}

	// 收集所有唯一的模型ID
	modelIdMap := make(map[string]bool)
	for _, dialog := range resp.Dialogs {
		modelIdMap[dialog.ModelId] = true
	}

	// 提取唯一模型ID列表
	var uniqueModelIds []string
	for modelId := range modelIdMap {
		uniqueModelIds = append(uniqueModelIds, modelId)
	}

	// 批量检查模型权限
	authorizedModelIds, _ := CheckModelUserPermission(ctx, userId, orgId, uniqueModelIds)

	// 创建授权模型ID的映射，用于快速查找
	authorizedModelMap := make(map[string]bool)
	for _, modelId := range authorizedModelIds {
		authorizedModelMap[modelId] = true
	}

	// 过滤出用户有权限的对话
	var dialogs []*response.ModelExperienceDialog
	for _, dialog := range resp.Dialogs {
		if authorizedModelMap[dialog.ModelId] {
			dialogs = append(dialogs, toModelExperienceDialog(dialog))
		}
	}

	return &response.ListResult{
		List:  dialogs,
		Total: int64(len(resp.Dialogs)),
	}, nil
}

func DeleteModelExperienceDialog(ctx *gin.Context, userId, orgId, modelExperienceId string) error {
	_, err := model.DeleteModelExperienceDialog(ctx.Request.Context(), &model_service.ModelExperienceDialogReq{
		ModelExperienceId: modelExperienceId,
		UserId:            userId,
		OrgId:             orgId,
	})
	return err
}

func ListModelExperienceDialogRecords(ctx *gin.Context, userId, orgId string, req *request.ModelExperienceDialogRecordRequest) (*response.ListResult, error) {
	recordsResp, err := model.GetModelExperienceDialogRecords(ctx.Request.Context(), &model_service.GetModelExperienceDialogRecordsReq{
		UserId: userId,
		OrgId:  orgId,
		// 常规模型对话记录（非模型对比时），modelExperienceId非空，sessionId前端没传
		ModelExperienceId: req.ModelExperienceId,
		SessionId:         "",
	})
	if err != nil {
		return nil, err
	}
	var records []*response.ModelExperienceDialogRecord
	for _, record := range recordsResp.Records {
		records = append(records, &response.ModelExperienceDialogRecord{
			ModelExperienceId: record.ModelExperienceId,
			ModelId:           record.ModelId,
			SessionId:         record.SessionId,
			OriginalContent:   record.OriginalContent,
			ReasoningContent:  record.ReasoningContent,
			Role:              record.Role,
			RequestFiles:      toAssistantRequestFileFromModelExperienceFileInfo(record.FileInfo),
		})
	}
	return &response.ListResult{
		List:  records,
		Total: int64(len(records)),
	}, nil
}
func toModelExperienceDialog(dialog *model_service.ModelExperienceDialog) *response.ModelExperienceDialog {
	return &response.ModelExperienceDialog{
		ID:           dialog.ModelExperienceId,
		ModelId:      dialog.ModelId,
		SessionId:    dialog.SessionId,
		Title:        dialog.Title,
		ModelSetting: dialog.ModelSetting,
		CreatedAt:    dialog.CreatedAt,
		UpdatedAt:    dialog.UpdatedAt,
	}
}

// --- modelExperienceSensitiveService: 实现 chatService 接口，供 ProcessSensitiveWords 使用 ---
type modelExperienceSensitiveService struct{}

func (m *modelExperienceSensitiveService) serviceType() string {
	return "ModelExperience"
}

func (m *modelExperienceSensitiveService) parseContent(raw string) (id, content string) {
	// SSE 数据格式为 "data: {...}\n\n"，解析出 content + reasoning_content
	raw = strings.TrimPrefix(raw, "data:")
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[DONE]" {
		return "", ""
	}

	var resp mp_common.LLMResp
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", ""
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Delta != nil {
		delta := resp.Choices[0].Delta
		content = delta.Content
		if delta.ReasoningContent != nil && *delta.ReasoningContent != "" {
			content = content + *delta.ReasoningContent // 合并思维链和正文内容
		}
	}
	return resp.ID, content
}

func (m *modelExperienceSensitiveService) buildSensitiveResp(id, content string) []string {
	// 返回 OpenAI delta 格式的 SSE 数据（finish_reason: stop 表示流结束）
	resp := mp_common.LLMResp{
		ID: id,
		Choices: []mp_common.OpenAIRespChoice{
			{
				Index: 0,
				Delta: &mp_common.OpenAIMsg{
					Role:    mp_common.MsgRoleAssistant,
					Content: content,
				},
				FinishReason: "stop",
			},
		},
	}
	marshal, _ := json.Marshal(resp)
	return []string{"data: " + string(marshal) + "\n\n"}
}

// toJsonFromModelExperienceFileInfo 将 FileInfo 列表序列化为 JSON 字符串，FileUrl 转为相对路径存储
func toJsonFromModelExperienceFileInfo(files []request.ConversionStreamFile) string {
	if len(files) == 0 {
		return ""
	}
	// 复制切片，避免修改原始数据，并将 FileUrl 转为相对路径存储
	result := make([]request.ConversionStreamFile, len(files))
	for i, f := range files {
		result[i] = f
		// 将完整 URL 转为 bucket/objectName 格式存储
		bucketName, objectName, _ := minio_util.SplitMinioPath(f.FileUrl)
		if bucketName != "" && objectName != "" {
			result[i].FileUrl = bucketName + "/" + objectName
		}
	}
	data, err := json.Marshal(result)
	if err != nil {
		log.Errorf("toJsonFromModelExperienceFileInfo error: %v", err)
		return ""
	}
	return string(data)
}

// toAssistantRequestFileFromModelExperienceFileInfo 从 JSON 字符串解析并转换为响应格式
func toAssistantRequestFileFromModelExperienceFileInfo(fileInfoJson string) []response.AssistantRequestFile {
	files := replaceModelExperienceMinioUrl(fileInfoJson)
	if len(files) == 0 {
		return nil
	}
	result := make([]response.AssistantRequestFile, 0, len(files))
	for _, f := range files {
		result = append(result, response.AssistantRequestFile{
			FileName: f.FileName,
			FileSize: f.FileSize,
			FileUrl:  f.FileUrl,
		})
	}
	return result
}

// replaceModelExperienceMinioUrl 从 JSON 字符串解析并替换相对路径为对外下载 URL
func replaceModelExperienceMinioUrl(fileInfoJson string) []request.ConversionStreamFile {
	if fileInfoJson == "" {
		return nil
	}
	var files []request.ConversionStreamFile
	if err := json.Unmarshal([]byte(fileInfoJson), &files); err != nil {
		log.Errorf("replaceModelExperienceMinioUrl unmarshal error: %v", err)
		return nil
	}
	for i := range files {
		files[i].FileUrl, _ = url.JoinPath(config.Cfg().Minio.DownloadURL, files[i].FileUrl)
	}
	return files
}

// checkModelExperienceModelVisionSupport 检查模型是否支持图文问答
func checkModelExperienceModelVisionSupport(modelInfo *model_service.ModelInfo) (bool, error) {
	allModelTags, err := getModelAllTags(modelInfo)
	if err != nil {
		return false, err
	}
	for _, tag := range allModelTags {
		if tag.Text == mp_common.TagVisionSupport {
			return true, nil
		}
	}
	return false, nil
}

// getMaxImageSize 从模型配置中获取 maxImageSize（单位：MB）
// validateImageFiles 校验图片文件
func validateModelExperienceImageFiles(ctx context.Context, modelInfo *model_service.ModelInfo, files []request.ConversionStreamFile) error {
	// 获取模型的 maxImageSize 配置（单位：MB），0表示不限制
	maxImageSize := int64(0)
	if modelInfo.ProviderConfig != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(modelInfo.ProviderConfig), &config); err == nil {
			if size, ok := config["maxImageSize"].(float64); ok {
				maxImageSize = int64(size)
			}
		}
	}

	const mb = 1024 * 1024
	supportedImageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".tiff"}
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.FileName))
		if !slices.Contains(supportedImageExts, ext) {
			return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("unsupported image format: %s, supported formats: jpg, jpeg, png, gif, bmp, webp, svg, tiff", ext))
		}
		// 使用前端传入的 FileSize 进行校验
		if maxImageSize > 0 && file.FileSize > maxImageSize*mb {
			return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("image size exceeds limit: %s, max size: %dMB", file.FileName, maxImageSize))
		}
	}
	return nil
}

// buildMultimodalContent 构建多模态消息内容
// 如果没有文件，返回纯文本；否则返回 OpenAI 多模态格式（图片转 base64）
func buildModelExperienceMultimodalContent(ctx context.Context, content string, fileInfo []request.ConversionStreamFile) interface{} {
	if len(fileInfo) == 0 {
		return content
	}

	// OpenAI 多模态格式
	contents := make([]map[string]interface{}, 0)

	// 先添加文本
	if content != "" {
		contents = append(contents, map[string]interface{}{
			"type": "text",
			"text": content,
		})
	}

	// 添加文件（图片转 base64）
	for _, file := range fileInfo {
		if file.FileUrl == "" {
			continue
		}
		// 从 minio 下载图片并转成 base64
		_, base64StrWithPrefix, err := minio_util.MinioUrlToBase64(ctx, file.FileUrl)
		if err != nil {
			log.Errorf("buildMultimodalContent: failed to convert image to base64: %v, url: %s", err, file.FileUrl)
			continue
		}
		contents = append(contents, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": base64StrWithPrefix,
			},
		})
	}

	// 如果没有有效的文件，返回纯文本
	if len(contents) == 0 || (len(contents) == 1 && content != "") {
		return content
	}
	return contents
}

// modelExperienceSSECompactProcessor 构造 SSE 消息合并处理器
// 用于合并 LLM 流式响应的 delta 消息，减少内存存储
func modelExperienceSSECompactProcessor() func(currentMsg *sse_model.Message, lastMsg *sse_model.Message) (bool, *sse_model.Message) {
	return func(currentMsg *sse_model.Message, lastMsg *sse_model.Message) (bool, *sse_model.Message) {
		// 无效消息（如SSE空行）不存储，避免阻断合并链
		currentMsgDataTrim := strings.TrimSpace(currentMsg.Data.(string))
		if currentMsgDataTrim == "" {
			if lastMsg != nil {
				return false, lastMsg // currentMsg不入缓存，lastMsg覆盖最新数据
			}
			return true, currentMsg
		}
		// 判断是否需要合并
		noneProcess, lastMsgData, currentMsgData := modelExperienceNoneCompactMessage(currentMsg, lastMsg)
		if noneProcess {
			return true, currentMsg
		}
		//开始合并
		compact := lastMsgData.Compact(currentMsgData)
		if compact != nil { //合并成功
			resp, err := response.MarshalModelExperienceResp(compact)
			if err != nil {
				log.Errorf("marshal agent resp error %v", err)
				return true, currentMsg
			}
			lastMsg.Data = resp
			return false, lastMsg
		}
		return true, currentMsg
	}
}

// modelExperienceNoneCompactMessage 判断是否需要合并
func modelExperienceNoneCompactMessage(currentMsg *sse_model.Message, lastMsg *sse_model.Message) (bool, *response.ModelExperienceResp, *response.ModelExperienceResp) {

	lastMsgData, err1 := response.UnmarshalModelExperienceResp(lastMsg.Data.(string))
	if err1 != nil {
		log.Errorf("unmarshal modelExperience resp %s error %v", lastMsg.Data.(string), err1)
		return true, nil, nil
	}
	currentMsgData, err2 := response.UnmarshalModelExperienceResp(currentMsg.Data.(string))
	if err2 != nil {
		log.Errorf("unmarshal modelExperience resp %s error %v", currentMsg.Data.(string), err2)
		return true, nil, nil
	}
	return false, lastMsgData, currentMsgData
}
