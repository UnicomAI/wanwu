package openapi

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// 知识问答 OpenAPI —— 应用管理。对外用 uuid 指代应用，知识问答的 uuid 即 ragId。

// CreateRag
//
//	@Tags			openapi
//	@Summary		创建知识问答OpenAPI
//	@Description	创建一个知识问答应用（草稿态）。返回的 uuid 用于后续配置、发布与问答接口。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPICreateRagRequest	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.OpenAPICreateRagResponse}
//	@Router			/rag [post]
func CreateRag(ctx *gin.Context) {
	var req request.OpenAPICreateRagRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	resp, err := service.CreateRag(ctx, userID, orgID, req.AppBriefConfig)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, response.OpenAPICreateRagResponse{UUID: resp.RagID}, nil)
}

// UpdateRag
//
//	@Tags			openapi
//	@Summary		更新知识问答基本信息OpenAPI
//	@Description	修改名称、描述、图标等基本信息（模型、知识库等高级配置请使用 /rag/config）。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIUpdateRagRequest	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/rag [put]
func UpdateRag(ctx *gin.Context) {
	var req request.OpenAPIUpdateRagRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	err := service.UpdateRag(ctx, request.RagUpdateReq{
		RagID:          req.UUID,
		AppBriefConfig: req.AppBriefConfig,
	}, userID, orgID)
	gin_util.Response(ctx, nil, err)
}

// DeleteRag
//
//	@Tags			openapi
//	@Summary		删除知识问答OpenAPI
//	@Description	删除指定知识问答，同时删除其已发布版本。删除后不可恢复，请谨慎操作。
//	@Produce		json
//	@Param			uuid	query		string	true	"知识问答UUID"
//	@Success		200		{object}	response.Response
//	@Router			/rag [delete]
func DeleteRag(ctx *gin.Context) {
	var req request.OpenAPIRagDeleteRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	err := service.DeleteAppSpaceApp(ctx, userID, orgID, req.UUID, constant.AppTypeRag)
	gin_util.Response(ctx, nil, err)
}

// ListRags
//
//	@Tags			openapi
//	@Summary		知识问答列表OpenAPI
//	@Description	获取当前 API Key 所属用户下的全部知识问答（含草稿与已发布），可按名称模糊筛选。返回的 uuid 可用于其他知识问答接口。
//	@Produce		json
//	@Param			name	query		string	false	"按名称模糊筛选（可选）"
//	@Success		200		{object}	response.Response{data=response.OpenAPIRagListResponse}
//	@Router			/rag/list [get]
func ListRags(ctx *gin.Context) {
	var req request.OpenAPIRagListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	resp, err := service.GetRagListForOpenAPI(ctx, userID, orgID, req.Name)
	gin_util.Response(ctx, resp, err)
}

// GetRagInfo
//
//	@Tags			openapi
//	@Summary		获取知识问答详情OpenAPI
//	@Description	获取知识问答详情，通过 published 参数控制返回草稿态或已发布态配置。返回中的模型 id 均为模型UUID。
//	@Produce		json
//	@Param			uuid		query		string	true	"知识问答UUID"
//	@Param			published	query		bool	false	"true=已发布态，false=草稿态（默认）"
//	@Success		200			{object}	response.Response{data=response.RagInfo}
//	@Router			/rag/info [get]
func GetRagInfo(ctx *gin.Context) {
	var req request.OpenAPIGetRagInfoRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	if req.Published {
		if err := service.CheckOpenAPIAccess(ctx, req.UUID, constant.AppTypeRag, userID, orgID); err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
	}
	resp, err := service.GetRag(ctx, userID, orgID, request.RagReq{RagID: req.UUID}, req.Published)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 对外统一用模型UUID
	for _, cfg := range []*request.AppModelConfig{&resp.ModelConfig, &resp.RerankConfig, &resp.QARerankConfig} {
		if cfg.ModelId == "" {
			continue
		}
		if uuid, convErr := service.GetModelUuidById(ctx, cfg.ModelId); convErr == nil {
			cfg.ModelId = uuid
		}
	}
	gin_util.Response(ctx, resp, nil)
}

// UpdateRagConfig
//
//	@Tags			openapi
//	@Summary		更新知识问答配置OpenAPI
//	@Description	更新模型、知识库、问答库、安全护栏等配置。除 uuid 外均为可选，只传要改的字段，未传的沿用当前草稿配置；模型 id 传模型UUID。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIRagConfigUpdateRequest	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/rag/config [put]
func UpdateRagConfig(ctx *gin.Context) {
	var req request.OpenAPIRagConfigUpdateRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	cfg, err := buildRagConfigFromCurrent(ctx, userID, orgID, req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, nil, service.UpdateRagConfig(ctx, userID, orgID, *cfg))
}

// buildRagConfigFromCurrent 以当前草稿配置为底，覆盖请求里传了的字段
func buildRagConfigFromCurrent(ctx *gin.Context, userID, orgID string, req request.OpenAPIRagConfigUpdateRequest) (*request.RagConfig, error) {
	cur, err := service.GetRag(ctx, userID, orgID, request.RagReq{RagID: req.UUID}, false)
	if err != nil {
		return nil, err
	}
	cfg := &request.RagConfig{
		RagID:                 req.UUID,
		ModelConfig:           &cur.ModelConfig,
		RerankConfig:          &cur.RerankConfig,
		QARerankConfig:        &cur.QARerankConfig,
		KnowledgeBaseConfig:   &cur.KnowledgeBaseConfig,
		QAKnowledgeBaseConfig: &cur.QAKnowledgeBaseConfig,
		SafetyConfig:          &cur.SafetyConfig,
		VisionConfig:          &cur.VisionConfig,
		RecommendQuestion:     cur.RecommendQuestion,
	}
	// 三个模型配置传的是模型UUID，落库要换回内部 id
	if cfg.ModelConfig, err = resolveOpenAPIModelConfig(ctx, req.ModelConfig, cfg.ModelConfig); err != nil {
		return nil, err
	}
	if cfg.RerankConfig, err = resolveOpenAPIModelConfig(ctx, req.RerankConfig, cfg.RerankConfig); err != nil {
		return nil, err
	}
	if cfg.QARerankConfig, err = resolveOpenAPIModelConfig(ctx, req.QARerankConfig, cfg.QARerankConfig); err != nil {
		return nil, err
	}
	if req.KnowledgeBaseConfig != nil {
		cfg.KnowledgeBaseConfig = req.KnowledgeBaseConfig
	}
	if req.QAKnowledgeBaseConfig != nil {
		cfg.QAKnowledgeBaseConfig = req.QAKnowledgeBaseConfig
	}
	if req.SafetyConfig != nil {
		cfg.SafetyConfig = req.SafetyConfig
	}
	if req.VisionConfig != nil {
		cfg.VisionConfig = req.VisionConfig
	}
	if req.RecommendQuestion != nil {
		cfg.RecommendQuestion = req.RecommendQuestion
	}
	return cfg, nil
}

// resolveOpenAPIModelConfig 请求未传该模型配置时沿用 fallback；传了则把模型UUID换成内部模型id，
// 并按模型信息补齐调用方没传的 provider/model/modelType——config 参数要靠 provider+modelType 才解析得出来
func resolveOpenAPIModelConfig(ctx *gin.Context, in, fallback *request.AppModelConfig) (*request.AppModelConfig, error) {
	if in == nil {
		return fallback, nil
	}
	cfg := *in
	if cfg.ModelId != "" {
		modelInfo, err := service.GetModelByUuid(ctx, cfg.ModelId)
		if err != nil {
			return nil, err
		}
		cfg.ModelId = modelInfo.GetModelId()
		if cfg.Provider == "" {
			cfg.Provider = modelInfo.GetProvider()
		}
		if cfg.Model == "" {
			cfg.Model = modelInfo.GetModel()
		}
		if cfg.ModelType == "" {
			cfg.ModelType = modelInfo.GetModelType()
		}
	}
	return &cfg, nil
}

// PublishRag
//
//	@Tags			openapi
//	@Summary		发布知识问答OpenAPI
//	@Description	发布知识问答，发布后可通过 /rag/chat 进行正式问答。发布前需先配置对话模型。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIRagPublishRequest	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/rag/publish [post]
func PublishRag(ctx *gin.Context) {
	var req request.OpenAPIRagPublishRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	err := service.PublishApp(ctx, userID, orgID, request.PublishAppRequest{
		AppId:       req.UUID,
		AppType:     constant.AppTypeRag,
		Version:     req.Version,
		Desc:        req.Desc,
		PublishType: req.PublishType,
	})
	gin_util.Response(ctx, nil, err)
}
