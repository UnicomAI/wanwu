package middleware

import (
	"encoding/json"

	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	utils "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

const (
	traceStatisticSourceKey             = "trace_statistic_source"
	traceStatisticModuleKey             = "trace_statistic_module"
	traceStatisticModuleResourceIDKey   = "trace_statistic_module_resource_id"
	traceStatisticModuleResourceTypeKey = "trace_statistic_module_resource_type"
	traceStatisticModuleCreatorUserKey  = "trace_statistic_module_creator_user_id"
	traceStatisticModuleCreatorOrgKey   = "trace_statistic_module_creator_org_id"
	traceStatisticClientIDKey           = "trace_statistic_client_id"
	traceStatisticAPIKeyKey             = "trace_statistic_api_key"
	traceStatisticUrlIDKey              = "trace_statistic_url_id"
)

// TraceWeb 设置 web source + module；moduleCreator 由 opts 显式指定。
// 仅写 gin.Context，持久化由全局 TraceStatistic post 中间件统一完成。
func TraceWeb(module string, opts ...TraceStatisticOption) gin.HandlerFunc {
	return traceWeb(constant.BizSourceWeb, module, nil, opts)
}

// TraceOpenAPI 设置 openapi source + 明文 apiKey；moduleCreator 由 opts 显式指定。
func TraceOpenAPI(module string, opts ...TraceStatisticOption) gin.HandlerFunc {
	return traceWeb(constant.BizSourceOpenAPI, module, []gin.HandlerFunc{TraceAPIKeyPlain()}, opts)
}

// TraceOpenUrl 设置 webURL source + clientId + 从 suffix 解析 app 创建人与资源。
func TraceOpenUrl(module string, opts ...TraceStatisticOption) gin.HandlerFunc {
	return traceWeb(constant.BizSourceWebUrl, module, []gin.HandlerFunc{TraceClientID(), TraceOpenUrlApp()}, opts)
}

// traceWeb 构造一个合并 handler：依次执行 pre（设 source/apiKey/clientId/查 app 等），
// 再设置 module 与 opts；持久化由全局 TraceStatistic post 中间件统一完成。
func traceWeb(source, module string, pre []gin.HandlerFunc, opts []TraceStatisticOption) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(traceStatisticSourceKey, source)
		for _, h := range pre {
			h(ctx)
			if ctx.IsAborted() {
				return
			}
		}
		ctx.Set(traceStatisticModuleKey, module)
		for _, opt := range opts {
			opt(ctx)
			if ctx.IsAborted() {
				return
			}
		}
		ctx.Next()
	}
}

// TraceOpenUrlApp 根据 :suffix 路径参数写入 app 创建人与 appId。
func TraceOpenUrlApp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		suffix := ctx.Param("suffix")
		if suffix == "" {
			return
		}
		appUrlInfo, err := service.LookupAppUrlBySuffix(ctx.Request.Context(), suffix)
		if err != nil {
			log.Errorf("trace openurl app suffix %q failed: %v", suffix, err)
			gin_util.Response(ctx, nil, err)
			ctx.Abort()
			return
		}
		setModuleCreator(ctx, appUrlInfo.UserId, appUrlInfo.OrgId)
		setModuleResource(ctx, appUrlInfo.AppId, appUrlInfo.AppType)
		ctx.Set(traceStatisticUrlIDKey, appUrlInfo.UrlId)
	}
}

// TraceStatisticOption 配置 TraceStatistic 中间件。
type TraceStatisticOption func(*gin.Context)

// WithModuleCreatorFromContext 使用当前登录用户作为 module 创建人。
func WithModuleCreatorFromContext() TraceStatisticOption {
	return func(ctx *gin.Context) {
		userID, orgID, err := getUserInfo(ctx)
		if err != nil {
			abortTraceStatistic(ctx, err.Error())
			return
		}
		setModuleCreator(ctx, userID, orgID)
	}
}

// WithDraftAppResource 草稿态 app 对话：写入 appId/appType，moduleCreator 取当前调用人（不查发布表）。
func WithDraftAppResource(appType, resourceField string) TraceStatisticOption {
	return func(ctx *gin.Context) {
		appID := getFieldValue(ctx, resourceField)
		if appID == "" {
			abortTraceStatistic(ctx, "app resource id is required")
			return
		}
		setModuleResource(ctx, appID, appType)
		userID, orgID, err := getUserInfo(ctx)
		if err != nil {
			abortTraceStatistic(ctx, err.Error())
			return
		}
		setModuleCreator(ctx, userID, orgID)
	}
}

// WithAppResource 从请求字段读取 appId，并查已发布 app 创建人作为 moduleCreator。
func WithAppResource(appType, resourceField string) TraceStatisticOption {
	return func(ctx *gin.Context) {
		appID := getFieldValue(ctx, resourceField)
		if appID == "" {
			abortTraceStatistic(ctx, "app resource id is required")
			return
		}
		setModuleResource(ctx, appID, appType)
		userID, orgID, err := service.LookupAppCreator(ctx.Request.Context(), appID, appType)
		if err != nil {
			log.Errorf("trace app resource %q type %q creator lookup failed: %v", appID, appType, err)
			abortTraceStatistic(ctx, err.Error())
			return
		}
		setModuleCreator(ctx, userID, orgID)
	}
}

// WithKnowledgeResource 根据 knowledgeId 写入统计维度：AppID=knowledgeId，moduleCreator=知识库创建人。
func WithKnowledgeResource(knowledgeIDField string) TraceStatisticOption {
	return func(ctx *gin.Context) {
		knowledgeID := getFieldValue(ctx, knowledgeIDField)
		if knowledgeID == "" {
			abortTraceStatistic(ctx, "knowledge id is required")
			return
		}
		setModuleResource(ctx, knowledgeID, constant.BizModuleResourceKnowledge)
		creatorUserID, creatorOrgID, err := service.LookupKnowledgeCreator(ctx.Request.Context(), knowledgeID)
		if err != nil {
			log.Errorf("trace knowledge resource %q creator lookup failed: %v", knowledgeID, err)
			abortTraceStatistic(ctx, err.Error())
			return
		}
		setModuleCreator(ctx, creatorUserID, creatorOrgID)
	}
}

// WithModelResource 根据 modelId 查询模型导入人作为 moduleCreator。
// model 模块无 app 资源维度，不写入 resourceID（应用维度记板块级空 appId，由 recordModelStatistic 双写）。
func WithModelResource(modelIDField string) TraceStatisticOption {
	return func(ctx *gin.Context) {
		modelID := getFieldValue(ctx, modelIDField)
		if modelID == "" {
			abortTraceStatistic(ctx, "model id is required")
			return
		}
		creatorUserID, creatorOrgID, err := service.LookupModelCreator(ctx.Request.Context(), modelID)
		if err != nil {
			log.Errorf("trace model resource %q creator lookup failed: %v", modelID, err)
			abortTraceStatistic(ctx, err.Error())
			return
		}
		setModuleCreator(ctx, creatorUserID, creatorOrgID)
	}
}

// WithKnowledgeHitResource 根据 knowledgeList 写入统计维度：
// - 有效知识库 ID 数 = 1：AppID=knowledgeId，AppType=knowledge，moduleCreator=知识库创建人（web/openapi 同逻辑）
// - 有效知识库 ID 数 > 1：暂不写入统计维度（见 FIXME）
// - 无有效 ID：不在支持范围内，请求 abort
func WithKnowledgeHitResource() TraceStatisticOption {
	return func(ctx *gin.Context) {
		knowledgeIDs := extractKnowledgeHitIDs(ctx)
		if len(knowledgeIDs) == 0 {
			abortTraceStatistic(ctx, "knowledge id is required")
			return
		}
		if len(knowledgeIDs) > 1 {
			// FIXME: 多知识库命中暂不落统计（无单一 AppID）；后续按多资源维度补齐后再记。
			// 清掉 TraceWeb 已写入的 module，避免半成品 Trace 导致模型统计反复 Parse 失败刷日志。
			ctx.Set(traceStatisticModuleKey, "")
			return
		}

		knowledgeID := knowledgeIDs[0]
		// AppType=knowledge，供模型统计与应用统计落库（经 Trace.moduleResourceType）
		setModuleResource(ctx, knowledgeID, constant.BizModuleResourceKnowledge)
		creatorUserID, creatorOrgID, err := service.LookupKnowledgeCreator(ctx.Request.Context(), knowledgeID)
		if err != nil {
			log.Errorf("trace knowledge resource %q creator lookup failed: %v", knowledgeID, err)
			abortTraceStatistic(ctx, err.Error())
			return
		}
		setModuleCreator(ctx, creatorUserID, creatorOrgID)
	}
}

// WithOpenAPIDraftAgentResource 草稿态 OpenAPI 智能体对话：uuid 解析为 assistantId，moduleCreator 取 API Key 所属用户。
func WithOpenAPIDraftAgentResource(uuidField string) TraceStatisticOption {
	return func(ctx *gin.Context) {
		if _, ok := resolveAssistantID(ctx, uuidField); !ok {
			return
		}
		userID, orgID, err := getUserInfo(ctx)
		if err != nil {
			abortTraceStatistic(ctx, err.Error())
			return
		}
		setModuleCreator(ctx, userID, orgID)
	}
}

// WithOpenAPIAgentResource 解析 OpenAPI uuid 为 assistant uuid，并查已发布 app 创建人。
func WithOpenAPIAgentResource(uuidField string) TraceStatisticOption {
	return func(ctx *gin.Context) {
		appID, ok := resolveAssistantID(ctx, uuidField)
		if !ok {
			return
		}
		userID, orgID, err := service.LookupAppCreator(ctx.Request.Context(), appID, constant.AppTypeAgent)
		if err != nil {
			log.Errorf("trace openapi agent %q creator lookup failed: %v", appID, err)
			abortTraceStatistic(ctx, err.Error())
			return
		}
		setModuleCreator(ctx, userID, orgID)
	}
}

// resolveAssistantID 将 OpenAPI uuid 解析为内部 assistant uuid，并写入资源维度。
func resolveAssistantID(ctx *gin.Context, uuidField string) (appID string, ok bool) {
	uuid := getFieldValue(ctx, uuidField)
	if uuid == "" {
		abortTraceStatistic(ctx, "agent uuid is required")
		return "", false
	}
	// appID, err := service.GetAssistantIdByUuid(ctx.Request.Context(), uuid)
	// if err != nil {
	// 	log.Errorf("trace agent uuid %q resolve failed: %v", uuid, err)
	// 	abortTraceStatistic(ctx, err.Error())
	// 	return "", false
	// }
	appID = uuid
	setModuleResource(ctx, appID, constant.AppTypeAgent)
	return appID, true
}

// TraceStatistic 全局 post 中间件：在所有路由中间件执行完后统一 persistTraceInfo 一次。
func TraceStatistic(ctx *gin.Context) {
	defer utils.PrintPanicStackWithCall(func(panicOccur bool, recoverError error) {
		if panicOccur {
			log.Errorf("trace statistic panic %v", recoverError)
		}
		ctx.Next()
	})
	if err := persistTraceInfo(ctx); err != nil {
		log.Errorf("persist trace info failed: %v", err)
	}
}

// TraceClientID 读取 X-Client-ID 写入 Context（作为 TraceOpenUrl 链内的 pre，不单独持久化）。
func TraceClientID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientID := ctx.GetHeader(gin_util.X_CLIENT_ID)
		if clientID == "" {
			clientID = ctx.GetString(gin_util.X_CLIENT_ID)
		}
		if clientID != "" {
			ctx.Set(traceStatisticClientIDKey, clientID)
		}
	}
}

// TraceAPIKeyPlain 将 OpenAPI 明文 apiKey 写入 Context（作为 TraceOpenAPI 链内的 pre，不单独持久化）。
func TraceAPIKeyPlain() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if key, err := getApiKey(ctx); err == nil && key != "" {
			ctx.Set(traceStatisticAPIKeyKey, key)
		}
	}
}

func setModuleCreator(ctx *gin.Context, userID, orgID string) {
	ctx.Set(traceStatisticModuleCreatorUserKey, userID)
	ctx.Set(traceStatisticModuleCreatorOrgKey, orgID)
}

func setModuleResource(ctx *gin.Context, appID, appType string) {
	ctx.Set(traceStatisticModuleResourceIDKey, appID)
	ctx.Set(traceStatisticModuleResourceTypeKey, appType)
}

// persistTraceInfo 读 Redis → 无则 buildTraceInfo → merge 统计维度 → 写回，单次 GET+SET。
// callback 无 TraceWeb 时：依赖上游（workflow test_run / 知识库导入等）已写入的 TraceInfo；
// Redis 尚无记录则跳过，避免用空 TraceInfo 占坑导致模型统计 Parse 失败。
func persistTraceInfo(ctx *gin.Context) error {
	reqCtx := ctx.Request.Context()
	if !trace_util.IsTraceContextValid(reqCtx) {
		return nil
	}
	traceID := trace_util.GetTraceID(reqCtx)

	info, err := trace_util.GetTraceUser(reqCtx)
	if err != nil {
		return err
	}
	hasGinStat := ctx.GetString(traceStatisticSourceKey) != "" ||
		ctx.GetString(traceStatisticModuleKey) != ""
	if info == nil {
		if !hasGinStat {
			return nil
		}
		info = buildTraceInfo(ctx, traceID)
	}

	mergeGinTraceStatistic(ctx, info)

	data, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	return redis.OP().Cli().Set(reqCtx, trace_util.TraceUserKey(traceID), data, buildKeyTimeout(ctx)).Err()
}

func mergeGinTraceStatistic(ctx *gin.Context, info *common.TraceInfo) {
	if info.TraceExtra == nil {
		info.TraceExtra = make(map[string]string)
	}

	setExtra := func(key, ginKey string) {
		if v := ctx.GetString(ginKey); v != "" {
			info.TraceExtra[key] = v
		}
	}

	setExtra(trace_util.TraceExtraSource, traceStatisticSourceKey)
	setExtra(trace_util.TraceExtraModule, traceStatisticModuleKey)
	setExtra(trace_util.TraceExtraModuleResourceID, traceStatisticModuleResourceIDKey)
	setExtra(trace_util.TraceExtraModuleResourceType, traceStatisticModuleResourceTypeKey)
	setExtra(trace_util.TraceExtraModuleCreatorUser, traceStatisticModuleCreatorUserKey)
	setExtra(trace_util.TraceExtraModuleCreatorOrg, traceStatisticModuleCreatorOrgKey)
	setExtra(trace_util.TraceExtraClientID, traceStatisticClientIDKey)

	ensureTraceUser := func() *common.TraceUser {
		if info.TraceUser == nil {
			info.TraceUser = &common.TraceUser{}
		}
		return info.TraceUser
	}
	if apiKey := ctx.GetString(traceStatisticAPIKeyKey); apiKey != "" {
		ensureTraceUser().ApiKey = apiKey
	}
	if apiKeyId := ctx.GetString(gin_util.API_KEY_ID); apiKeyId != "" {
		ensureTraceUser().ApiKeyId = apiKeyId
	}
	if urlID := ctx.GetString(traceStatisticUrlIDKey); urlID != "" {
		ensureTraceUser().UrlId = urlID
	}

	// module 为 app 类资源时同步 traceApp
	if appID := info.TraceExtra[trace_util.TraceExtraModuleResourceID]; appID != "" &&
		IsAppStatisticModule(info.TraceExtra[trace_util.TraceExtraModule]) {
		if info.TraceApp == nil {
			info.TraceApp = &common.TraceApp{}
		}
		info.TraceApp.AppId = appID
		info.TraceApp.AppType = info.TraceExtra[trace_util.TraceExtraModuleResourceType]
	}
}

func abortTraceStatistic(ctx *gin.Context, msg string) {
	log.Errorf("trace statistic setup failed: %s", msg)
	gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_trace_statistic", msg))
	ctx.Abort()
}

func extractKnowledgeHitIDs(ctx *gin.Context) []string {
	body, err := requestBody(ctx)
	if err != nil || body == "" {
		return nil
	}
	var req struct {
		KnowledgeList []struct {
			ID string `json:"id"`
		} `json:"knowledgeList"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return nil
	}
	ids := make([]string, 0, len(req.KnowledgeList))
	for _, item := range req.KnowledgeList {
		if item.ID != "" {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

// IsAppStatisticModule 判断 module 维度是否为 app 类资源（assistant/rag/workflow），
// 用于统计 V2 中 traceApp 维度的同步。
func IsAppStatisticModule(module string) bool {
	switch module {
	case constant.BizModuleAppAgent, constant.BizModuleAppRag, constant.BizModuleAppWorkflow:
		return true
	}
	return false
}
