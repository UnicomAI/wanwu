package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"

	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// 定义校验函数类型
type ModelValidator func(ctx *gin.Context, modelInfo *model_service.ModelInfo) error

// 校验器注册表
var validators = sync.OnceValue(func() map[string]ModelValidator {
	return map[string]ModelValidator{
		mp.ModelTypeLLM:            ValidateLLMModel,
		mp.ModelTypeTextRerank:     ValidateRerankModel,
		mp.ModelTypeMultiRerank:    ValidateMultiRerankModel,
		mp.ModelTypeTextEmbedding:  ValidateEmbeddingModel,
		mp.ModelTypeMultiEmbedding: ValidateMultiEmbeddingModel,
		mp.ModelTypeOcr:            ValidateOcrModel,
		mp.ModelTypeGui:            ValidateGuideModel,
		mp.ModelTypePdfParser:      ValidatePdfParserModel,
		mp.ModelTypeSyncAsr:        ValidateSyncAsrModel,
		mp.ModelTypeText2Image:     ValidateText2ImageModel,
	}
})

// 统一校验入口
func ValidateModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	validator, exists := validators()[strings.ToLower(modelInfo.ModelType)]
	if !exists {
		return fmt.Errorf("unsupported model type: %s", modelInfo.ModelType)
	}
	return validator(ctx, modelInfo)
}

func ValidateLLMModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	// mock  request
	stream := false
	var result map[string]interface{}
	err = json.Unmarshal([]byte(modelInfo.ProviderConfig), &result)
	if err != nil {
		return err
	}

	toolCallFlag := false // ToolCall 校验标识
	fc, ok := result["functionCalling"].(string)
	if ok && mp_common.FCType(fc) == mp_common.FCTypeToolCall {
		toolCallFlag = true
	}

	visionSupportFlag := false // VisionSupport 校验标识
	vs, ok := result["visionSupport"].(string)
	if ok && mp_common.VSType(vs) == mp_common.VSTypeSupport {
		visionSupportFlag = true
	}

	// 工具调用校验
	if toolCallFlag {
		reqTool := &mp_common.LLMReq{
			Model: modelInfo.Model,
			Messages: []mp_common.OpenAIReqMsg{
				{
					Role:    mp_common.MsgRoleUser,
					Content: "What time is it in Beijing now?", // 工具调用专属Content
				},
			},
			Stream: &stream,
		}
		tools := []mp_common.OpenAITool{
			{
				Type: mp_common.ToolTypeFunction,
				Function: &mp_common.OpenAIFunction{
					Name:        "get_current_time",
					Description: "It's very useful when you want to know the current time in Beijing.",
				},
			},
		}
		reqTool.Tools = tools
		// 执行工具调用校验
		llmReqTool, err := iLLM.NewReq(reqTool)
		if err != nil {
			return err
		}
		respTool, _, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReqTool)
		if err != nil {
			return fmt.Errorf("toolcall validation failed: %v, maybe model does not support toolcall functionality", err)
		}
		openAIRespTool, ok := respTool.ConvertResp()
		if !ok {
			return fmt.Errorf("toolcall validation: invalid response format")
		}
		if len(openAIRespTool.Choices) == 0 || openAIRespTool.Choices[0].Message.ToolCalls == nil {
			return fmt.Errorf("model does not support toolcall functionality")
		}
		if openAIRespTool.Choices[0].FinishReason != "tool_calls" {
			return fmt.Errorf("model does not support toolcall functionality: finish_reason is %s, expected tool_calls", openAIRespTool.Choices[0].FinishReason)
		}
		// 打印工具调用日志
		data, _ := json.MarshalIndent(openAIRespTool.Choices[0].Message.ToolCalls, "", "  ")
		log.Debugf("tool call: %v", string(data))
	}

	_, base64StrWithPrefix, err := util.File2Base64(config.Cfg().Model.PngTestFilePath, "")
	if err != nil {
		return err
	}
	// 视觉支持校验（独立请求，用专属Content和配置）
	if visionSupportFlag {
		// 视觉支持专属req：Content为图片+“这里有什么字”，无Tools配置
		reqVision := &mp_common.LLMReq{
			Model: modelInfo.Model,
			Messages: []mp_common.OpenAIReqMsg{
				{
					Role: mp_common.MsgRoleUser,
					Content: []map[string]interface{}{ // 视觉支持专属Content
						{
							"type": "image_url",
							"image_url": map[string]string{
								"url": base64StrWithPrefix,
							},
						},
						{
							"type": "text",
							"text": "描述一下图片",
						},
					},
				},
			},
			Stream: &stream,
		}
		// 执行视觉支持校验
		llmReq, err := iLLM.NewReq(reqVision)
		if err != nil {
			return err
		}

		resp, _, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReq)
		if err != nil {
			return fmt.Errorf("vision validation failed: %v, maybe model does not support vision functionality", err)
		}

		openAIResp, ok := resp.ConvertResp()
		if !ok {
			return fmt.Errorf("vision validation: invalid response format")
		}

		if len(openAIResp.Choices) == 0 || openAIResp.Choices[0].Message == nil {
			return fmt.Errorf("vision validation: empty response")
		}

		content := openAIResp.Choices[0].Message.Content
		// 检查响应内容是否包含 "GPU"，确认模型能识别图片中的文字内容
		if content == "" || !strings.Contains(content, "GPU") {
			return fmt.Errorf("model does not support vision functionality: response indicates image was not processed")
		}

		return nil
	}

	if !toolCallFlag && !visionSupportFlag {
		// 执行基础校验
		reqBase := &mp_common.LLMReq{
			Model: modelInfo.Model,
			Messages: []mp_common.OpenAIReqMsg{
				{
					Role:    mp_common.MsgRoleUser,
					Content: "ping",
				},
			},
			Stream: &stream,
		}
		llmReqBase, err := iLLM.NewReq(reqBase)
		if err != nil {
			return fmt.Errorf("base llm validation failed: %v", err)
		}
		baseResp, _, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReqBase)
		if err != nil {
			return fmt.Errorf("base llm chat failed: %v, maybe model is unavailable", err)
		}
		_, ok = baseResp.ConvertResp()
		if !ok {
			return fmt.Errorf("invalid response format")
		}
	}
	return nil
}

func ValidateEmbeddingModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	embedding, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iEmbedding, ok := embedding.(mp.IEmbedding)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	// mock  request
	req := &mp_common.EmbeddingReq{
		Model: modelInfo.Model,
		Input: []string{"你好"},
	}
	embeddingReq, err := iEmbedding.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iEmbedding.Embeddings(ctx.Request.Context(), embeddingReq)
	if err != nil {
		{
			return fmt.Errorf("model API call failed: %v", err)
		}
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func ValidateMultiEmbeddingModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	embedding, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iMultiModalEmbedding, ok := embedding.(mp.IMultiModalEmbedding)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	_, base64StrWithPrefix, err := util.File2Base64(config.Cfg().Model.PngTestFilePath, "")
	if err != nil {
		return err
	}
	// mock  request
	textHello := "你好"
	req := &mp_common.MultiModalEmbeddingReq{
		Model: modelInfo.Model,
		Input: []mp_common.MultiInput{
			{
				Text: &textHello,
			},
			{
				Image: &base64StrWithPrefix,
			},
		},
	}
	multiModalEmbeddingReq, err := iMultiModalEmbedding.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iMultiModalEmbedding.MultiModalEmbeddings(ctx.Request.Context(), multiModalEmbeddingReq)
	if err != nil {
		{
			return fmt.Errorf("model API call failed: %v", err)
		}
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func ValidateRerankModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	rerank, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iRerank, ok := rerank.(mp.IRerank)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	// mock  request
	req := &mp_common.TextRerankReq{
		Model: modelInfo.Model,
		Query: "乌萨奇",
		Documents: []string{
			"乌萨奇",
			"尖尖我噶奶～",
		},
	}
	rerankReq, err := iRerank.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iRerank.Rerank(ctx.Request.Context(), rerankReq)
	if err != nil {
		return fmt.Errorf("model API call failed: %v", err)
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func ValidateMultiRerankModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	rerank, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iRerank, ok := rerank.(mp.IMultiModalRerank)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	base64Str, base64StrWithPrefix, err := util.File2Base64(config.Cfg().Model.PngTestFilePath, "")
	if err != nil {
		return err
	}
	// jina 传参为不带前缀base64
	urlData := base64StrWithPrefix
	if modelInfo.Provider == mp.ProviderJina {
		urlData = base64Str
	}
	// mock  request
	textBeiJi := "北极"
	req := &mp_common.MultiModalRerankReq{
		Model: modelInfo.Model,
		Query: "企鹅",
		Documents: []mp_common.MultiDocument{
			{
				Text: &textBeiJi,
			},
			{
				Image: &urlData,
			},
		},
	}

	rerankReq, err := iRerank.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iRerank.MultiModalRerank(ctx.Request.Context(), rerankReq)
	if err != nil {
		return fmt.Errorf("model API call failed: %v", err)
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func ValidateOcrModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	ocr, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iOcr, ok := ocr.(mp.IOcr)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	// mock request：读取测试文件转 base64
	fileData, err := os.ReadFile(config.Cfg().Model.PngTestFilePath)
	if err != nil {
		return fmt.Errorf("read test file failed: %v", err)
	}
	_, base64StrWithPrefix, err := util.FileData2Base64(fileData, "")
	if err != nil {
		return fmt.Errorf("file to base64 failed: %v", err)
	}
	req := &mp_common.OcrReq{
		FileData: &base64StrWithPrefix,
		FileName: "test.png",
		Model:    &modelInfo.Model,
	}
	ocrReq, err := iOcr.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iOcr.Ocr(ctx, ocrReq)
	if err != nil {
		return fmt.Errorf("model API call failed: %v", err)
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func ValidatePdfParserModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	pdfParser, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iPdfParser, ok := pdfParser.(mp.IPdfParser)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	// mock  request
	file, err := os.Open(config.Cfg().Model.PdfTestFilePath)
	if err != nil {
		return fmt.Errorf("open file failed: %v", err)
	}
	defer func() { _ = file.Close() }()

	// 创建内存缓冲区
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建表单文件字段
	part, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		return fmt.Errorf("create form file failed: %v", err)
	}

	// 复制文件内容
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file content failed: %v", err)
	}
	_ = writer.Close()

	// 模拟HTTP请求
	mockReq, _ := http.NewRequest("POST", "", body)
	mockReq.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Request = mockReq
	// 获取FileHeader对象
	_, fileH, err := ctx.Request.FormFile("file")
	if err != nil {
		return fmt.Errorf("get file header failed: %v", err)
	}
	req := &mp_common.PdfParserReq{
		Files:    fileH,
		FileName: "test.pdf",
	}
	pdfParserReq, err := iPdfParser.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iPdfParser.PdfParser(ctx, pdfParserReq)
	if err != nil {
		return fmt.Errorf("model API call failed: %v", err)
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func ValidateSyncAsrModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	asr, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iAsr, ok := asr.(mp.ISyncAsr)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	// mock  request
	req, err := getSyncAsrReqByProvider(ctx, modelInfo)
	if err != nil {
		return err
	}
	asrReq, err := iAsr.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iAsr.SyncAsr(ctx, asrReq)
	if err != nil {
		return fmt.Errorf("model API call failed: %v", err)
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func ValidateGuideModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	gui, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iGui, ok := gui.(mp.IGui)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	// mock  request
	// 读取图片文件
	imageFile := config.Cfg().Model.PngTestFilePath
	imageBytes, err := os.ReadFile(imageFile)
	if err != nil {
		return fmt.Errorf("ReadFile file failed: %v", err)
	}

	// 转换为base64字符串
	imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)
	height, width := 931, 144
	req := &mp_common.GuiReq{
		Algo:                    "gui_agent_v1",
		Platform:                "Mobile",
		CurrentScreenshot:       "data:image/jpeg;base64," + imageBase64,
		CurrentScreenshotHeight: height,
		CurrentScreenshotWidth:  width,
		Task:                    "点击屏幕以开始",
		History:                 []string{},
	}
	guiReq, err := iGui.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iGui.Gui(ctx.Request.Context(), guiReq)
	if err != nil {
		return fmt.Errorf("model API call failed: %v", err)
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func ValidateText2ImageModel(ctx *gin.Context, modelInfo *model_service.ModelInfo) error {
	text2Image, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iText2Image, ok := text2Image.(mp.IText2Image)
	if !ok {
		return fmt.Errorf("invalid provider")
	}
	// mock  request
	advOpt := mp_common.AdvancedOptJson{
		Height:             512,
		Width:              512,
		NumImagesPerPrompt: 1,
		Style:              "摄影",
	}
	advOptByte, _ := json.Marshal(advOpt)

	req := &mp_common.Text2ImageReq{
		Prompt:         "小丑",
		ResponseFormat: "url",
		AdvancedOpt:    string(advOptByte),
	}
	text2ImageReq, err := iText2Image.NewReq(req)
	if err != nil {
		return err
	}
	resp, err := iText2Image.Text2Image(ctx, text2ImageReq)
	if err != nil {
		return fmt.Errorf("model API call failed: %v", err)
	}
	_, ok = resp.ConvertResp()
	if !ok {
		return fmt.Errorf("invalid response format")
	}
	return nil
}

func getSyncAsrReqByProvider(ctx *gin.Context, modelInfo *model_service.ModelInfo) (*mp_common.SyncAsrReq, error) {
	_, base64StrWithPrefix, err := util.File2Base64(config.Cfg().Model.
		AsrTestFilePath, "")
	if err != nil {
		return nil, fmt.Errorf("file_upload_file_2_base64: %v", err)
	}
	req := &mp_common.SyncAsrReq{
		Model: modelInfo.Model,
		Messages: []mp_common.SyncAsrReqMsg{
			{
				Role: "user",
				Content: []mp_common.SyncAsrReqC{
					{
						Type: mp_common.MultiModalTypeMinioUrl,
						Audio: mp_common.SyncAsrAudio{
							Data: base64StrWithPrefix,
						},
					},
				},
			},
		},
	}
	switch modelInfo.Provider {
	case mp.ProviderYuanJing:
		req.Messages[0].Content[0].Audio.FileName = "test.wav"
	}
	return req, nil
}

func validateModelThinking(ctx context.Context, iLLM mp.ILLM, model string, stream *bool, enableThinking bool) error {
	temp := enableThinking
	reqThinking := &mp_common.LLMReq{
		Model: model,
		Messages: []mp_common.OpenAIReqMsg{
			{
				Role:    mp_common.MsgRoleUser,
				Content: "say hi",
			},
		},
		Stream:         stream,
		EnableThinking: &temp,
	}
	llmReqThinking, err := iLLM.NewReq(reqThinking)
	if err != nil {
		return err
	}

	respThinking, _, err := iLLM.ChatCompletions(ctx, llmReqThinking)
	if err != nil {
		if enableThinking {
			return fmt.Errorf("thinking validation (enable) failed: %v, maybe model does not support thinking functionality", err)
		}
		return fmt.Errorf("thinking validation (disable) failed: %v", err)
	}
	openAIRespThinking, ok := respThinking.ConvertResp()
	if !ok {
		return fmt.Errorf("thinking validation: invalid response format")
	}

	if enableThinking {
		if len(openAIRespThinking.Choices) == 0 || openAIRespThinking.Choices[0].Message == nil ||
			openAIRespThinking.Choices[0].Message.ReasoningContent == nil || *openAIRespThinking.Choices[0].Message.ReasoningContent == "" {
			return fmt.Errorf("model does not support thinking functionality when enable_thinking=true")
		}
	} else {
		if len(openAIRespThinking.Choices) > 0 && openAIRespThinking.Choices[0].Message != nil &&
			openAIRespThinking.Choices[0].Message.ReasoningContent != nil && *openAIRespThinking.Choices[0].Message.ReasoningContent != "" {
			return fmt.Errorf("model does not support turning off thinking functionality")
		}
	}
	return nil
}

func ValidateModelThinking(ctx *gin.Context, userId, orgId string, req *request.ImportOrUpdateModelRequest) error {
	modelInfo, err := parseImportAndUpdateClientReq(userId, orgId, req)
	if err != nil {
		return err
	}

	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return err
	}
	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		return fmt.Errorf("invalid provider")
	}

	stream := false
	if err := validateModelThinking(ctx.Request.Context(), iLLM, modelInfo.Model, &stream, true); err != nil {
		return err
	}
	if err := validateModelThinking(ctx.Request.Context(), iLLM, modelInfo.Model, &stream, false); err != nil {
		return err
	}
	return nil
}
