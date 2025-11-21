package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/gin-gonic/gin"
)

// ImportModel
//
//	@Tags			model
//	@Summary model import
//	@Description Import of third-party models
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.ImportOrUpdateModelRequest true "Import model information"
//	@Success		200		{object}	response.Response
//	@Router			/model [post]
func ImportModel(ctx *gin.Context) {
	var req request.ImportOrUpdateModelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ImportModel(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateModel
//
//	@Tags		model
//	@Summary Import model updates
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.ImportOrUpdateModelRequest true "Model change information"
//	@Success	200		{object}	response.Response
//	@Router		/model [put]
func UpdateModel(ctx *gin.Context) {
	var req request.ImportOrUpdateModelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateModel(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteModel
//
//	@Tags		model
//	@Summary Import model deletion
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.DeleteModelRequest true "Model delete key"
//	@Success	200		{object}	response.Response
//	@Router		/model [delete]
func DeleteModel(ctx *gin.Context) {
	var req request.DeleteModelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteModel(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// GetModel
//
//	@Tags		model
//	@Summary ‌Query a single model
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.GetModelRequest true "Model ID"
//	@Success	200		{object}	response.Response{data=response.ModelInfo}
//	@Router		/model [get]
func GetModel(ctx *gin.Context) {
	var req request.GetModelRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetModel(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ListModels
//
//	@Tags		model
//	@Summary Import model list
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param modelType query string false "model type" Enums(llm,embedding,rerank)
//	@Param provider query string false "model provider"
//	@Param displayName query string false "Model display name"
//	@Param isActive query string false "Enabled status (true: enabled)"
//	@Success	200			{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/list [get]
func ListModels(ctx *gin.Context) {
	var req request.ListModelsRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ListModels(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ChangeModelStatus
//
//	@Tags		model
//	@Summary model enable/disable
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param data body request.ModelStatusRequest true "Enable/disable model information"
//	@Success	200		{object}	response.Response
//	@Router		/model/status [put]
func ChangeModelStatus(ctx *gin.Context) {
	var req request.ModelStatusRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ChangeModelStatus(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// ListLlmModels
//
//	@Tags		model
//	@Summary llm model list
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/llm [get]
func ListLlmModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeLLM,
	})
	gin_util.Response(ctx, resp, err)
}

// ListRerankModels
//
//	@Tags		model
//	@Summary rerank model list
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/rerank [get]
func ListRerankModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeRerank,
	})
	gin_util.Response(ctx, resp, err)
}

// ListEmbeddingModels
//
//	@Tags		model
//	@Summary embedding model list
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/embedding [get]
func ListEmbeddingModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeEmbedding,
	})
	gin_util.Response(ctx, resp, err)
}

// ListOcrModels
//
//	@Tags		model
//	@Summary ocr model list
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/ocr [get]
func ListOcrModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeOcr,
	})
	gin_util.Response(ctx, resp, err)
}

// ListPdfParserModels
//
//	@Tags		model
//	@Summary pdf document parsing model list
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/pdf-parser [get]
func ListPdfParserModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypePdfParser,
	})
	gin_util.Response(ctx, resp, err)
}

// ListGuiModels
//
//	@Tags		model
//	@Summary gui model list
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/gui [get]
func ListGuiModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeGui,
	})
	gin_util.Response(ctx, resp, err)
}
