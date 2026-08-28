package middleware

import (
	"encoding/json"
	"time"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	RecordStreamType    = "stream"
	RecordNonStreamType = "non_stream"
	RecordFromReq       = "req"
)

// APIKeyRecord 记录 API Key 调用的中间件
// 需要在 AuthOpenAPIKey 中间件之后使用
func APIKeyRecord(StreamType string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 检查是否有 API Key ID（由 AuthOpenAPIKey 设置）
		apiKeyID := ctx.GetString(gin_util.API_KEY_ID)
		if apiKeyID == "" {
			ctx.Next()
			return
		}

		startTime := time.Now()
		var reqBody string
		var isStream bool
		// 仅 JSON 读取 body：requestBody 会 ReadAll 且按 JSON 解析，multipart 会耗尽显存 body，
		// 导致后续 MultipartForm / ShouldBind 报 multipart: NextPart: EOF（与 Record 中间件一致）。
		if ctx.ContentType() == gin.MIMEJSON {
			reqBody, _ = requestBody(ctx)
		}
		switch StreamType {
		case RecordStreamType:
			isStream = true
		case RecordNonStreamType:
			isStream = false
		default:
			// 通过请求体检测是否为流式请求
			if ctx.ContentType() == gin.MIMEJSON {
				isStream = detectStreamRequest(reqBody)
			}
		}

		ctx.Next()

		// 获取 HTTP 状态码
		httpStatus := ctx.Writer.Status()
		if httpStatus == 0 {
			httpStatus = ctx.GetInt(gin_util.STATUS)
		}
		if httpStatus == 0 {
			httpStatus = 200
		}

		// 计算耗时
		var streamCosts, nonStreamCosts int64
		if isStream {
			// 流式请求：从 ctx 获取首 token 时延；未设置则保持 0（不用总耗时冒充 TTFT）
			streamCosts = ctx.GetInt64(gin_util.FIRST_RESP_LATENCY)
		} else {
			nonStreamCosts = time.Since(startTime).Milliseconds()
		}

		// 获取响应体（非流式请求）
		var responseBody string
		if !isStream {
			responseBody = ctx.GetString(gin_util.RESULT)
		}

		// 构建方法路径
		methodPath := ctx.Request.Method + "-" + ctx.Request.URL.Path

		// 记录调用（使用 DetachContext 防止客户端断开后 context 被取消导致统计丢失）
		detachedCtx := trace_util.DetachContext(ctx.Request.Context())
		go func() {
			defer util.PrintPanicStack()
			service.RecordAPIKeyCall(detachedCtx,
				ctx.GetString(gin_util.USER_ID),
				ctx.GetString(gin_util.X_ORG_ID),
				apiKeyID,
				methodPath,
				startTime.UnixMilli(),
				int64(httpStatus),
				isStream,
				streamCosts,
				nonStreamCosts,
				reqBody,
				responseBody,
			)
		}()
	}
}

// detectStreamRequest 从请求体检测是否为流式请求
// 只有 stream=true 才认为是流式请求
func detectStreamRequest(body string) bool {
	if body == "" {
		return false
	}
	var req struct {
		Stream bool `json:"stream"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return false
	}
	return req.Stream
}
