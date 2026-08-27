package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

//	@title		AI Agent Productivity Platform - Open API
//	@version	v0.0.1

//	@BasePath	/openapi/v1

// CreateAgent
//
//	@Tags			openapi
//	@Summary		创建智能体OpenAPI
//	@Description	创建智能体OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPICreateAgentRequest	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.OpenAPICreateAgentResponse}
//	@Router			/agent [post]
func CreateAgent(ctx *gin.Context) {
	var req request.OpenAPICreateAgentRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	assistantCreateResp, err := service.AssistantCreate(ctx, userID, orgID, request.AssistantCreateReq(req))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, response.OpenAPICreateAgentResponse{UUID: assistantCreateResp.AssistantId}, nil)
}

// CreateAgentConversation
//
//	@Tags			openapi
//	@Summary		智能体创建对话OpenAPI
//	@Description	智能体创建对话OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentCreateConversationRequest	true	"请求参数"
//	@Success		400		{object}	response.Response{data=response.OpenAPIAgentCreateConversationResponse}
//	@Router			/agent/conversation [post]
func CreateAgentConversation(ctx *gin.Context) {
	var req request.OpenAPIAgentCreateConversationRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	if err := service.CheckOpenAPIAccess(ctx, req.UUID, constant.AppTypeAgent, userID, orgID); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// OpenAPI 创建的对话和 web 已发布对话归到同一类型（type=published）
	resp, err := service.ConversationCreate(ctx, userID, orgID, request.ConversationCreateRequest{
		AssistantId: req.UUID,
		Prompt:      req.Title,
	}, constant.ConversationTypePublished)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, response.OpenAPIAgentCreateConversationResponse{ConversationID: resp.ConversationId}, nil)
}

// ChatAgent
//
//	@Tags			openapi
//	@Summary		智能体对话OpenAPI
//	@Description	智能体对话OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentChatRequest	true	"请求参数"
//	@Success		200		{object}	response.OpenAPIAgentChatResponse
//	@Success		400		{object}	response.Response
//	@Router			/agent/chat [post]
func ChatAgent(ctx *gin.Context) {
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	var req request.OpenAPIAgentChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	clientID := getClientID(ctx)
	//appID, err := service.GetAssistantIdByUuid(ctx, req.UUID)
	//if err != nil {
	//	gin_util.Response(ctx, nil, err)
	//	return
	//}
	if err := service.CheckOpenAPIAccess(ctx, req.UUID, constant.AppTypeAgent, userID, orgID); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 流式
	if req.Stream {
		if err := service.AssistantConversionStream(ctx, userID, orgID, clientID, request.ConversionStreamRequest{
			AssistantId:    req.UUID,
			ConversationId: req.ConversationID,
			Prompt:         req.Query,
			FileInfo:       req.FileInfo,
		}, true, constant.BizSourceOpenAPI); err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
		if qs := service.GenAgentRecommendQuestions(ctx, userID, orgID, req.UUID, req.ConversationID, req.Query, true); len(qs) > 0 {
			if b, err := json.Marshal(map[string][]string{"recommend_questions": qs}); err == nil {
				_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", string(b))
				ctx.Writer.Flush()
			}
		}
		return
	}
	// 非流式
	startTime := time.Now()
	chatCh, err := service.CallAssistantConversationStream(ctx, userID, orgID, clientID, request.ConversionStreamRequest{
		AssistantId:    req.UUID,
		ConversationId: req.ConversationID,
		Prompt:         req.Query,
		FileInfo:       req.FileInfo,
	}, true)
	if err != nil {
		statusCode, failureReason := service.GrpcErrorToHTTPStatus(err)
		go func() {
			defer util.PrintPanicStack()
			service.RecordAppStatistic(detachedCtx, userID, orgID, req.UUID, constant.AppTypeAgent, "",
				statusCode, failureReason, false, 0, 0, constant.BizSourceOpenAPI, service.MarshalStatisticBody(req), "", req.Query, "")
		}()
		gin_util.Response(ctx, nil, err)
		return
	}
	var output strings.Builder
	resp := &response.OpenAPIAgentChatResponse{}
	for chat := range chatCh {
		// 注意这里智能体的原始流式返回没有data:前缀
		if strings.TrimSpace(chat) == "" {
			continue
		}
		curr := &response.OpenAPIAgentChatResponse{}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(chat, "data:")), curr); err != nil {
			log.Errorf("[Agent] %v conversation %v user %v org %v unmarshal %v err: %v", req.UUID, req.ConversationID, userID, orgID, err)
			continue
		}
		resp = curr
		output.WriteString(curr.Response)
	}
	resp.Response = output.String()
	resp.RecommendQuestions = service.GenAgentRecommendQuestions(ctx, userID, orgID, req.UUID, req.ConversationID, req.Query, true)
	costs := time.Since(startTime).Milliseconds()
	go func() {
		defer util.PrintPanicStack()
		service.RecordAppStatistic(detachedCtx, userID, orgID, req.UUID, constant.AppTypeAgent, "",
			200, "", false, 0, int64(costs), constant.BizSourceOpenAPI, service.MarshalStatisticBody(req), service.MarshalStatisticBody(resp), req.Query, resp.Response)
	}()
	b, _ := json.Marshal(resp)
	status := http.StatusOK
	ctx.Set(gin_util.STATUS, status)
	ctx.Set(gin_util.RESULT, string(b))
	ctx.JSON(status, resp)
}

// ChatRag
//
//	@Tags			openapi
//	@Summary		文本问答OpenAPI
//	@Description	文本问答OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIRagChatRequest	true	"请求参数"
//	@Success		200		{object}	response.OpenAPIRagChatResponse
//	@Success		400		{object}	response.Response
//	@Router			/rag/chat [post]
func ChatRag(ctx *gin.Context) {
	var req request.OpenAPIRagChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)

	// OpenAPI 权限检查：仅允许调用自己创建的应用
	if err := service.CheckOpenAPIAccess(ctx, req.UUID, constant.AppTypeRag, userID, orgID); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// conversation_id 由调用方给，校验会话归属，避免写进他人会话
	if req.ConversationID != "" {
		if err := service.CheckRagConversationAccess(ctx, req.ConversationID, req.UUID, userID, orgID); err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
	}
	// 带文件提问的前置校验
	if len(req.FileInfo) > 0 {
		if err := service.CheckRagChatFileReady(ctx, userID, orgID, req.UUID, true); err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
	}

	// openapi 固定走 legacy 格式，不跟随 web 的 AG-UI 协议改造
	chatRagWithFormat(ctx, userID, orgID, request.ChatRagRequest{
		RagID:          req.UUID,
		Question:       req.Query,
		FileInfo:       req.FileInfo,
		ConversationID: req.ConversationID,
	}, req.Stream, true, service.MarshalStatisticBody(req))
}

// DraftChatAgent
//
//	@Tags			openapi
//	@Summary		智能体草稿态对话OpenAPI
//	@Description	智能体草稿态对话OpenAPI，基于草稿配置进行问答，不要求智能体已发布，计入统计
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentDraftChatRequest	true	"请求参数"
//	@Success		200		{object}	response.OpenAPIAgentChatResponse
//	@Success		400		{object}	response.Response
//	@Router			/agent/chat/draft [post]
func DraftChatAgent(ctx *gin.Context) {
	var req request.OpenAPIAgentDraftChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	clientID := getClientID(ctx)
	//appID, err := service.GetAssistantIdByUuid(ctx, req.UUID)
	//if err != nil {
	//	gin_util.Response(ctx, nil, err)
	//	return
	//}
	appID := req.UUID
	// 草稿态不要求已发布，所有权由 GetAssistantInfo(..., false) 的 Identity 兜底（只能调自己的）
	assistantInfo, err := service.GetAssistantInfo(ctx, userID, orgID, request.AssistantIdRequest{AssistantId: appID}, false)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if assistantInfo.Prologue == "" {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_agent_prologue_required"))
		return
	}
	if assistantInfo.ModelConfig.ModelId == "" {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_agent_model_required"))
		return
	}

	// 草稿态每个智能体仅维护一条会话
	// 如果调用方没传conversationId，这里 get-or-create 一条 ConversationTypeDraft 会话，确保下游 ES 能落历史。
	if req.ConversationID == "" {
		convResp, err := service.GetDraftConversationIdByAssistantID(ctx, userID, orgID, request.ConversationGetListRequest{
			AssistantId: appID,
		})
		if err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
		if convResp == nil {
			// 草稿尚未发起过对话：自动创建一条
			newConvResp, createErr := service.ConversationCreate(ctx, userID, orgID, request.ConversationCreateRequest{
				AssistantId: appID,
				Prompt:      req.Query,
			}, constant.ConversationTypeDraft)
			if createErr != nil {
				gin_util.Response(ctx, nil, createErr)
				return
			}
			req.ConversationID = newConvResp.ConversationId
		} else {
			req.ConversationID = convResp.ConversationId
		}
	}

	if err := service.AssistantConversionStream(ctx, userID, orgID, clientID, request.ConversionStreamRequest{
		AssistantId:    appID,
		ConversationId: req.ConversationID,
		Prompt:         req.Query,
		FileInfo:       req.FileInfo,
	}, false, constant.BizSourceOpenAPI); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if qs := service.GenAgentRecommendQuestions(ctx, userID, orgID, appID, req.ConversationID, req.Query, false); len(qs) > 0 {
		if b, err := json.Marshal(map[string][]string{"recommend_questions": qs}); err == nil {
			_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", string(b))
			ctx.Writer.Flush()
		}
	}
}

// PublishAgent
//
//	@Tags			openapi
//	@Summary		智能体发布OpenAPI
//	@Description	发布智能体，发布后可通过智能体对话接口进行正式问答
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentPublishRequest	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/agent/publish [post]
func PublishAgent(ctx *gin.Context) {
	var req request.OpenAPIAgentPublishRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	//appID, err := service.GetAssistantIdByUuid(ctx, req.AssistantUUID)
	//if err != nil {
	//	gin_util.Response(ctx, nil, err)
	//	return
	//}
	appID := req.AssistantUUID
	assistantInfo, err := service.GetAssistantInfo(ctx, userID, orgID, request.AssistantIdRequest{AssistantId: appID}, false)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if assistantInfo.Prologue == "" {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_agent_prologue_required"))
		return
	}
	if assistantInfo.ModelConfig.ModelId == "" {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_agent_model_required"))
		return
	}
	err = service.PublishApp(ctx, userID, orgID, request.PublishAppRequest{
		AppId:       appID,
		AppType:     constant.AppTypeAgent,
		Version:     req.Version,
		Desc:        req.Desc,
		PublishType: req.PublishType,
	})
	gin_util.Response(ctx, nil, err)
}

// GetAgentInfo
//
//	@Tags			openapi
//	@Summary		获取智能体详情OpenAPI
//	@Description	获取智能体详情，通过 published 参数控制返回草稿态或已发布态配置
//	@Accept			json
//	@Produce		json
//	@Param			uuid		query		string	true	"智能体UUID"
//	@Param			published	query		bool	false	"true=已发布态，false=草稿态（默认）"
//	@Success		200			{object}	response.Response{data=response.Assistant}
//	@Router			/agent/info [get]
func GetAgentInfo(ctx *gin.Context) {
	var req request.OpenAPIGetAgentInfoRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	if req.Published {
		if err := service.CheckOpenAPIAccess(ctx, req.UUID, constant.AppTypeAgent, userID, orgID); err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
	}
	resp, err := service.GetAssistantInfo(ctx, userID, orgID, request.AssistantIdRequest{AssistantId: req.UUID}, req.Published)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if resp.ModelConfig.ModelId != "" {
		if uuid, convErr := service.GetModelUuidById(ctx, resp.ModelConfig.ModelId); convErr == nil {
			resp.ModelConfig.ModelId = uuid
		}
	}
	if resp.RerankConfig.ModelId != "" {
		if uuid, convErr := service.GetModelUuidById(ctx, resp.RerankConfig.ModelId); convErr == nil {
			resp.RerankConfig.ModelId = uuid
		}
	}
	if resp.RecommendConfig.ModelConfig.ModelId != "" {
		if uuid, convErr := service.GetModelUuidById(ctx, resp.RecommendConfig.ModelConfig.ModelId); convErr == nil {
			resp.RecommendConfig.ModelConfig.ModelId = uuid
		}
	}
	gin_util.Response(ctx, resp, nil)
}

// UpdateAgentConfig
//
//	@Tags			openapi
//	@Summary		更新智能体配置OpenAPI
//	@Description	更新开场白、提示词、模型、知识库等配置。除 assistantUuid 外均为可选，只传要改的字段，未传的沿用当前草稿配置；模型 id 传模型UUID。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentConfigUpdateRequest	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/agent/config [put]
func UpdateAgentConfig(ctx *gin.Context) {
	var req request.OpenAPIAgentConfigUpdateRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	cfg, err := buildAgentConfigFromCurrent(ctx, userID, orgID, req.AssistantUUID, req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	_, err = service.AssistantConfigUpdate(ctx, userID, orgID, *cfg)
	gin_util.Response(ctx, nil, err)
}

// buildAgentConfigFromCurrent 以当前草稿配置为底，覆盖请求里传了的字段。
// 只回填下游会无条件覆盖的字段（三个文本字段，以及为 nil 时会连带清空 rerank 的 knowledgeBaseConfig），
// 其余子配置留 nil 由下游保持原值——回填要重新解析模型/知识库，读不到时会把配置抹平。
func buildAgentConfigFromCurrent(ctx *gin.Context, userID, orgID, assistantID string, req request.OpenAPIAgentConfigUpdateRequest) (*request.AssistantConfig, error) {
	cur, err := service.GetAssistantInfo(ctx, userID, orgID, request.AssistantIdRequest{AssistantId: assistantID}, false)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "agent config update: assistant not found")
	}
	cfg := &request.AssistantConfig{
		AssistantId:         assistantID,
		SafetyConfig:        req.SafetyConfig,
		VisionConfig:        req.VisionConfig,
		MemoryConfig:        req.MemoryConfig,
		KnowledgeBaseConfig: mergeAgentKnowledgeBaseConfig(req.KnowledgeBaseConfig, cur.KnowledgeBaseConfig),
	}
	mergeAgentTextConfig(cfg, cur, req)
	if err := resolveAgentModelConfigs(ctx, cfg, req); err != nil {
		return nil, err
	}
	return cfg, nil
}

// mergeAgentTextConfig 三个文本字段下游会无条件覆盖，请求没传就回填当前草稿值
func mergeAgentTextConfig(cfg *request.AssistantConfig, cur *response.Assistant, req request.OpenAPIAgentConfigUpdateRequest) {
	cfg.Prologue, cfg.Instructions, cfg.RecommendQuestion = cur.Prologue, cur.Instructions, cur.RecommendQuestion
	if req.Prologue != nil {
		cfg.Prologue = *req.Prologue
	}
	if req.Instructions != nil {
		cfg.Instructions = *req.Instructions
	}
	if req.RecommendQuestion != nil {
		cfg.RecommendQuestion = req.RecommendQuestion
	}
}

// resolveAgentModelConfigs 三处模型配置传的是模型UUID，落库要换回内部 id
func resolveAgentModelConfigs(ctx *gin.Context, cfg *request.AssistantConfig, req request.OpenAPIAgentConfigUpdateRequest) error {
	var err error
	fillDefaultLLMParams(req.ModelConfig)
	if cfg.ModelConfig, err = resolveOpenAPIModelConfig(ctx, req.ModelConfig, nil); err != nil {
		return err
	}
	if cfg.RerankConfig, err = resolveOpenAPIModelConfig(ctx, req.RerankConfig, nil); err != nil {
		return err
	}
	if req.RecommendConfig == nil {
		return nil
	}
	recCfg := *req.RecommendConfig
	fillDefaultLLMParams(&recCfg.ModelConfig)
	modelCfg, err := resolveOpenAPIModelConfig(ctx, &recCfg.ModelConfig, nil)
	if err != nil {
		return err
	}
	recCfg.ModelConfig = *modelCfg
	cfg.RecommendConfig = &recCfg
	return nil
}

// mergeAgentKnowledgeBaseConfig 请求没传知识库配置时沿用当前的——留 nil 下游会把 rerank 一并清空
func mergeAgentKnowledgeBaseConfig(in *request.AppKnowledgebaseConfig, cur request.AppKnowledgebaseConfig) *request.AppKnowledgebaseConfig {
	if in == nil {
		return &cur
	}
	kbCfg := *in
	kbCfg.Config = resolveKnowledgeParams(kbCfg.Config, cur.Config)
	return &kbCfg
}

// resolveKnowledgeParams 检索参数留空时沿用当前配置，都没有则套用页面侧默认值
func resolveKnowledgeParams(in, cur request.AppKnowledgebaseParams) request.AppKnowledgebaseParams {
	empty := request.AppKnowledgebaseParams{}
	if in != empty {
		return in
	}
	if cur != empty {
		return cur
	}
	return request.AppKnowledgebaseParams{MatchType: "mix", PriorityMatch: 1, Threshold: 0.4, TopK: 5}
}

// fillDefaultLLMParams 传了对话模型但没带推理参数时套用页面侧默认值
func fillDefaultLLMParams(cfg *request.AppModelConfig) {
	if cfg == nil || cfg.Config != nil || cfg.ModelType != "llm" {
		return
	}
	thinkingEnable := true
	cfg.Config = map[string]interface{}{
		"temperature":            0.7,
		"temperatureEnable":      true,
		"topP":                   1,
		"topPEnable":             true,
		"frequencyPenalty":       0,
		"frequencyPenaltyEnable": true,
		"presencePenalty":        0,
		"presencePenaltyEnable":  true,
		"maxTokens":              512,
		"maxTokensEnable":        true,
		"thinkingEnable":         &thinkingEnable,
	}
}

// --- internal ---

// 获取当前用户ID
func getUserID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.USER_ID)
}

// 获取当前组织ID
func getOrgID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.X_ORG_ID)
}

// 获取当前appID
func getAppID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.APP_ID)
}

// 获取当前clientID
func getClientID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.X_CLIENT_ID)
}
