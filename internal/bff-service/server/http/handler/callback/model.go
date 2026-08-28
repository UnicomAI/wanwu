package callback

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// GetModelById
//
//	@Tags		callback
//	@Summary	根据ModelId获取模型
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string	true	"模型ID"
//	@Success	200		{object}	response.Response{data=response.ModelInfo}
//	@Router		/model/{modelId} [get]
func GetModelById(ctx *gin.Context) {
	modelId := ctx.Param("modelId")
	resp, err := service.GetModelById(ctx, &request.GetModelRequest{
		BaseModelRequest: request.BaseModelRequest{ModelId: modelId}})
	// 替换callback返回的模型中的apiKey/endpointUrl信息
	if resp != nil && resp.Config != nil {
		cfg := make(map[string]interface{})
		b, err := json.Marshal(resp.Config)
		if err != nil {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v marshal config err: %v", modelId, err)))
			return
		}
		if err = json.Unmarshal(b, &cfg); err != nil {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v unmarshal config err: %v", modelId, err)))
			return
		}
		// 替换apiKey, endpointUrl
		cfg["apiKey"] = "useless-api-key"
		endpoint := mp.ToModelEndpoint(resp.ModelId, resp.Model)
		for k, v := range endpoint {
			if k == "model_url" {
				cfg["endpointUrl"] = v
				break
			}
		}
		// 替换Config
		resp.Config = cfg
	}
	gin_util.Response(ctx, resp, err)
}

// ModelChatCompletionsWithTrace 处理带 trace 上下文的模型代理请求。
//
// URL 中的 :traceId 和 :spanId 由 sandbox 内 opencode/eino-agent 调用时填入。
// TraceFromURLMiddleware 已在 otelgin 之前将此信息提取到 traceparent Header，
// 因此 otelgin 已创建了正确父子关系的 span，本 handler 只需直接委托。
//
//	@Tags		callback
//	@Summary	Model Chat Completions with Trace
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string				true	"模型ID"
//	@Param		traceId	path		string				true	"Trace ID"
//	@Param		spanId	path		string				true	"Span ID"
//	@Param		data	body		mp_common.LLMReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.LLMResp{}
//	@Router		/model/{modelId}/trace/{traceId}/span/{spanId}/chat/completions [post]
func ModelChatCompletionsWithTrace(ctx *gin.Context) {
	// trace 上下文的提取和 span 父子关系由 TraceFromURLMiddleware + otelgin 完成，
	// 此处不需要再做任何 trace 操作，直接委托给现有逻辑即可。
	ModelChatCompletions(ctx)
}

// ModelChatCompletions
//
//	@Tags		callback
//	@Summary	Model Chat Completions
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string				true	"模型ID"
//	@Param		data	body		mp_common.LLMReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.LLMResp{}
//	@Router		/model/{modelId}/chat/completions [post]
func ModelChatCompletions(ctx *gin.Context) {
	var data mp_common.LLMReq
	if !gin_util.Bind(ctx, &data) {
		return
	}
	// image url -> image base64
	if len(data.Messages) > 0 {
		for i := range data.Messages {
			if data.Messages[i].Role != mp_common.MsgRoleUser || data.Messages[i].Content == nil {
				continue
			}
			if _, ok := data.Messages[i].Content.(string); ok {
				continue
			}
			b, err := json.Marshal(data.Messages[i].Content)
			if err != nil {
				continue
			}
			var content []map[string]interface{}
			if err := json.Unmarshal(b, &content); err != nil {
				continue
			}
			var existImage bool
			for j, item := range content {
				var url map[string]string
				for k, v := range item {
					if k != "image_url" {
						continue
					}
					b, err := json.Marshal(v)
					if err != nil {
						continue
					}
					if err := json.Unmarshal(b, &url); err != nil {
						continue
					}
					var base64StrWithPrefix string
					for urlK, urlV := range url {
						if urlK == "url" {
							if err := url_util.ValidateURL(urlV); err != nil {
								log.Errorf("ModelChatCompletions invalid image URL: %v", err)
								continue
							}
							resp, err := http.Get(urlV)
							if err != nil {
								continue
							}
							defer func() { _ = resp.Body.Close() }()
							if resp.StatusCode != http.StatusOK {
								continue
							}
							body, err := io.ReadAll(resp.Body)
							if err != nil {
								continue
							}
							_, base64StrWithPrefix, err = util.FileData2Base64(body, "")
							if err != nil {
								continue
							}
							if base64StrWithPrefix != "" {
								url["url"] = base64StrWithPrefix
								break
							}
						}
					}
					if base64StrWithPrefix != "" {
						content[j]["image_url"] = url
						existImage = true
						break
					}
				}
			}
			if existImage {
				data.Messages[i].Content = content
			}
		}
	}
	service.ModelChatCompletions(ctx, ctx.Param("modelId"), &data, nil)
}

// ModelEmbeddings
//
//	@Tags		callback
//	@Summary	Model Embeddings
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string						true	"模型ID"
//	@Param		data	body		mp_common.EmbeddingReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.EmbeddingResp{}
//	@Router		/model/{modelId}/embeddings [post]
func ModelEmbeddings(ctx *gin.Context) {
	var data mp_common.EmbeddingReq
	if !gin_util.Bind(ctx, &data) {
		return
	}
	service.ModelEmbeddings(ctx, ctx.Param("modelId"), &data)
}

// ModelMultiModalEmbeddings
//
//	@Tags		callback
//	@Summary	Model MultiModal-Embeddings
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string								true	"模型ID"
//	@Param		data	body		mp_common.MultiModalEmbeddingReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.MultiModalEmbeddingResp{}
//	@Router		/model/{modelId}/multimodal-embeddings [post]
func ModelMultiModalEmbeddings(ctx *gin.Context) {
	var data mp_common.MultiModalEmbeddingReq
	if !gin_util.Bind(ctx, &data) {
		return
	}
	// file minio url -> file base64
	for i := range data.Input {
		item := &data.Input[i]
		if item.Image != nil && *item.Image != "" {
			_, base64StrWithPrefix, err := minio_util.MinioUrlToBase64(ctx, *item.Image)
			if err != nil {
				gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v image to base64 err: %v", data.Model, err)))
				return
			}
			item.Image = &base64StrWithPrefix
		}
		if item.Audio != nil && *item.Audio != "" {
			_, base64StrWithPrefix, err := minio_util.MinioUrlToBase64(ctx, *item.Audio)
			if err != nil {
				gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v audio to base64 err: %v", data.Model, err)))
				return
			}
			item.Audio = &base64StrWithPrefix
		}
		if item.Video != nil && *item.Video != "" {
			_, base64StrWithPrefix, err := minio_util.MinioUrlToBase64(ctx, *item.Video)
			if err != nil {
				gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v video to base64 err: %v", data.Model, err)))
				return
			}
			item.Video = &base64StrWithPrefix
		}
	}
	service.ModelMultiModalEmbeddings(ctx, ctx.Param("modelId"), &data)
}

// ModelTextRerank
//
//	@Tags		callback
//	@Summary	Model Rerank
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string						true	"模型ID"
//	@Param		data	body		mp_common.TextRerankReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.RerankResp{}
//	@Router		/model/{modelId}/rerank [post]
func ModelTextRerank(ctx *gin.Context) {
	var data mp_common.TextRerankReq
	if !gin_util.Bind(ctx, &data) {
		return
	}
	service.ModelTextRerank(ctx, ctx.Param("modelId"), &data)
}

// ModelMultiModalRerank
//
//	@Tags		callback
//	@Summary	Model MultiModal-Rerank
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string							true	"模型ID"
//	@Param		data	body		mp_common.MultiModalRerankReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.MultiModalRerankResp{}
//	@Router		/model/{modelId}/multimodal-rerank [post]
func ModelMultiModalRerank(ctx *gin.Context) {
	var data mp_common.MultiModalRerankReq
	if !gin_util.Bind(ctx, &data) {
		return
	}
	// file minio url -> file base64
	for i := range data.Documents {
		item := &data.Documents[i]
		if item.Image != nil && *item.Image != "" {
			_, base64StrWithPrefix, err := minio_util.MinioUrlToBase64(ctx, *item.Image)
			if err != nil {
				gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v image to base64 err: %v", data.Model, err)))
				return
			}
			item.Image = &base64StrWithPrefix
		}
	}
	switch queryVal := data.Query.(type) {
	case string:
	case map[string]interface{}:
		if imageUrl, ok := queryVal["image"].(string); ok && imageUrl != "" {
			_, base64StrWithPrefix, err := minio_util.MinioUrlToBase64(ctx, imageUrl)
			if err != nil {
				gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v query image to base64 err: %v", data.Model, err)))
				return
			}
			queryVal["image"] = base64StrWithPrefix
			data.Query = queryVal
		}
	default:
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("unsupported query type: %T, only support string or map[string]interface{}", queryVal)))
		return
	}
	service.ModelMultiModalRerank(ctx, ctx.Param("modelId"), &data)
}

// ModelOcr
//
//	@Tags		callback
//	@Summary	Model Ocr
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string				true	"模型ID"
//	@Param		data	body		mp_common.OcrReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.OcrResp{}
//	@Router		/model/{modelId}/ocr [post]
func ModelOcr(ctx *gin.Context) {
	var data mp_common.OcrReq
	if !gin_util.Bind(ctx, &data) {
		return
	}
	// url 转 base64（支持 minio URL）
	if data.Url != nil && *data.Url != "" {
		_, base64StrWithPrefix, err := minio_util.MinioUrlToBase64(ctx.Request.Context(), *data.Url)
		if err != nil {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model ocr url to base64 err: %v", err)))
			return
		}
		data.FileData = &base64StrWithPrefix
		data.Url = nil
	}
	service.ModelOcr(ctx, ctx.Param("modelId"), &data)
}

// ModelPdfParser
//
//	@Tags		callback
//	@Summary	Model PdfParser
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		modelId		path		string	true	"模型ID"
//	@Param		file		formData	file	true	"文件"
//	@Param		file_name	formData	string	true	"文件名"
//	@Success	200			{object}	mp_common.PdfParserResp{}
//	@Router		/model/{modelId}/pdf-parser [post]
func ModelPdfParser(ctx *gin.Context) {
	var data mp_common.PdfParserReq
	if !gin_util.BindForm(ctx, &data) {
		return
	}
	service.ModelPdfParser(ctx, ctx.Param("modelId"), &data)
}

// ModelGui
//
//	@Tags		callback
//	@Summary	Model Gui
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string				true	"模型ID"
//	@Param		data	body		mp_common.GuiReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.GuiResp{}
//	@Router		/model/{modelId}/gui [post]
func ModelGui(ctx *gin.Context) {
	var data mp_common.GuiReq
	if !gin_util.Bind(ctx, &data) {
		return
	}
	service.ModelGui(ctx, ctx.Param("modelId"), &data)
}

// ModelSyncAsr
//
//	@Tags		callback
//	@Summary	Model SyncAsr
//	@Accept		json
//	@Produce	json
//	@Param		modelId	path		string					true	"模型ID"
//	@Param		data	body		mp_common.SyncAsrReq{}	true	"请求参数"
//	@Success	200		{object}	mp_common.SyncAsrResp{}
//	@Router		/model/{modelId}/asr [post]
func ModelSyncAsr(ctx *gin.Context) {
	var data mp_common.SyncAsrReq
	if !gin_util.Bind(ctx, &data) {
		return
	}
	// file minio url -> file base64
	if len(data.Messages) > 0 {
		for i := range data.Messages {
			contentList := &data.Messages[i].Content
			for j := range *contentList {
				if (*contentList)[j].Type == mp_common.MultiModalTypeAudio {
					(*contentList)[j].Type = mp_common.MultiModalTypeMinioUrl
					minioFilePath := (*contentList)[j].Audio.Data
					_, base64StrWithPrefix, err := minio_util.MinioUrlToBase64(ctx.Request.Context(), minioFilePath)
					if err != nil {
						gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("minio_url %s to local file err: %v", minioFilePath, err)))
						return
					}
					(*contentList)[j].Audio.Data = base64StrWithPrefix
					(*contentList)[j].Audio.FileName = minio_util.GetFilenameFromMinioURL(minioFilePath)
				}
			}
		}
	}
	service.ModelSyncAsr(ctx, ctx.Param("modelId"), &data)
}
