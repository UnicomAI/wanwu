package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
)

const (
	RateLimitedI18nKey        = "bff_rate_limited_too_many_attempts"
	FailedWithAttemptsI18nKey = "bff_rate_limited_failed_with_attempts"
	DefaultKeyPrefix          = "rate_limiter"
	DefaultMaxAttempts        = 5
	DefaultCooldown           = 60 * time.Second
)

// LoginRateLimit 登录限流中间件（向后兼容）
// 基于 NewRateLimiter 的登录专用配置，保持与原 RateLimit 函数相同的行为。
var LoginRateLimit = NewRateLimiter(RateLimiterConfig{
	KeyPrefix:      "login_fail:",
	KeyFields:      []RateLimitKeyField{{Type: "body", Name: "username"}},
	MaxAttempts:    5,
	Cooldown:       60 * time.Second,
	ErrCodes:       []int64{int64(err_code.Code_IAMLogin)},
	ResetOnSuccess: true,
})

// RateLimitKeyField 限流 key 的字段来源
type RateLimitKeyField struct {
	// Type 字段来源类型："header" 从 HTTP Header 提取，"body" 从请求体 JSON 提取
	Type string
	// Name 字段名：
	//   - Type 为 "header" 时，为 HTTP Header 名称（如 "X-User-Id"）
	//   - Type 为 "body" 时，为 JSON 字段名，支持点号嵌套路径（如 "user.email"）
	Name string
}

// RateLimiterConfig 通用限流中间件的配置
type RateLimiterConfig struct {
	// KeyPrefix Redis key 前缀，例如 "login_fail:"
	KeyPrefix string

	// KeyFields 限流 key 的字段来源列表，用于构建限流 key。
	// 提取的值会自动转为小写，以确保限流 key 的大小写不敏感。
	// 如果为空，则仅使用 ClientIP 构建 key。
	KeyFields []RateLimitKeyField

	// MaxAttempts 最大连续失败次数，超过后触发限流锁定
	MaxAttempts int

	// Cooldown 限流计数器在 Redis 中的过期时间
	Cooldown time.Duration

	// ErrCodes 需要计入限流的响应错误码列表。
	// 如果为空，则所有响应都计入限流计数（不管成功还是失败）。
	// 指定具体错误码时，只有列表中的 code 才会计数，其他 code 直接放行。
	ErrCodes []int64

	// ResetOnSuccess 成功时（code == 0）是否清除计数器。
	// 适用于登录等场景：连续失败5次后，登录成功应清零。
	// 默认 false，成功时不影响计数。
	ResetOnSuccess bool
}

// responseBufferWriter 是一个 Gin ResponseWriter 的包装器，
// 用于缓冲响应内容，以便在需要时可以丢弃原始响应并重写。
type responseBufferWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func newResponseBufferWriter(w gin.ResponseWriter) *responseBufferWriter {
	return &responseBufferWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		status:         http.StatusOK,
	}
}

func (w *responseBufferWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *responseBufferWriter) WriteHeader(code int) {
	w.status = code
}

func (w *responseBufferWriter) WriteHeaderNow() {
	// 不立即写入 header，缓冲起来
}

func (w *responseBufferWriter) Status() int {
	return w.status
}

func (w *responseBufferWriter) Size() int {
	return w.body.Len()
}

func (w *responseBufferWriter) Written() bool {
	return w.body.Len() > 0
}

func (w *responseBufferWriter) Flush() {
	// no-op: 保持响应在缓冲区中，直到 flushToReal() 决定是写入还是丢弃
}

// flushToReal 将缓冲的响应写入真实的 ResponseWriter
func (w *responseBufferWriter) flushToReal() {
	w.ResponseWriter.WriteHeader(w.status)
	_, _ = w.ResponseWriter.Write(w.body.Bytes())
}

// NewRateLimiter 创建通用限流中间件
// 工作流程：
//  1. 前置：从请求体提取配置字段，构建 Redis key，检查是否已被限流
//  2. 缓冲响应：使用 responseBufferWriter 拦截 handler 的响应
//  3. 执行业务 handler（ctx.Next()）
//  4. 后置：读取缓冲的响应结果，判断是否计数
//     - ErrCodes 匹配：递增计数，达到上限时返回锁定消息
//     - ErrCodes 不匹配：直接放行，不影响计数
func NewRateLimiter(cfg RateLimiterConfig) gin.HandlerFunc {
	// 设置默认值
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = DefaultKeyPrefix
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = DefaultMaxAttempts
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = DefaultCooldown
	}

	return func(ctx *gin.Context) {
		// 1. 从请求体提取配置字段
		fields := extractKeyFields(ctx, cfg.KeyFields)
		if len(cfg.KeyFields) > 0 && len(fields) == 0 {
			// 配置了字段但无法提取，放行
			ctx.Next()
			return
		}

		// 2. 构建 Redis key
		key := buildRateLimitKey(cfg, fields, ctx.ClientIP())

		// 3. 前置检查：是否已被限流
		if checkRateLimitPre(ctx, cfg, key) {
			return // 已限流，直接返回
		}

		// 4. 缓冲响应，拦截 handler 写入的内容
		bufferWriter := newResponseBufferWriter(ctx.Writer)
		ctx.Writer = bufferWriter

		// 5. 执行业务 handler
		ctx.Next()

		// 6. 后置处理：检查响应结果
		handlePostCheck(ctx, cfg, key, bufferWriter)
	}
}

// extractKeyFields 从请求中提取配置的字段值
// RateLimitKeyField.Type 为 "header" 时从 HTTP Header 提取，为 "body" 时从请求体 JSON 提取。
// 提取的值自动转为小写，以确保限流 key 的大小写不敏感
func extractKeyFields(ctx *gin.Context, fields []RateLimitKeyField) map[string]string {
	if len(fields) == 0 {
		return nil
	}

	// 收集需要从 body 提取的字段，body 只解析一次
	var bodyFieldNames []string
	for _, f := range fields {
		if f.Type == "body" {
			bodyFieldNames = append(bodyFieldNames, f.Name)
		}
	}

	var paramsMap map[string]interface{}
	if len(bodyFieldNames) > 0 && ctx.ContentType() == gin.MIMEJSON {
		body, err := requestBody(ctx)
		if err == nil && body != "" {
			_ = json.Unmarshal([]byte(body), &paramsMap)
		}
	}

	result := make(map[string]string, len(fields))
	for _, f := range fields {
		var value string
		switch f.Type {
		case "header":
			// 从 HTTP Header 中提取
			value = ctx.GetHeader(f.Name)
		case "body":
			// 从请求体 JSON 中提取
			if paramsMap != nil {
				val, ok := getNestedValue(paramsMap, f.Name)
				if ok && val != nil {
					if strVal, ok := val.(string); ok && strVal != "" {
						value = strVal
					}
				}
			}
		}
		if value != "" {
			result[f.Name] = strings.ToLower(value)
		}
	}
	return result
}

// buildRateLimitKey 构建完整的 Redis key
func buildRateLimitKey(cfg RateLimiterConfig, fields map[string]string, ip string) string {
	// 默认拼接：按 KeyFields 顺序取字段值 + IP，以 ":" 分隔
	parts := make([]string, 0, len(cfg.KeyFields)+1)
	for _, f := range cfg.KeyFields {
		if v, ok := fields[f.Name]; ok && v != "" {
			parts = append(parts, v)
		}
	}
	parts = append(parts, ip)
	return cfg.KeyPrefix + strings.Join(parts, ":")
}

// checkRateLimitPre 前置检查：读取 Redis 计数器，如果达到限流上限则返回错误并终止请求
func checkRateLimitPre(ctx *gin.Context, cfg RateLimiterConfig, key string) bool {
	countStr, err := redis.Sys().Get(ctx.Request.Context(), key)
	if err != nil {
		// Redis 错误，fail-open（放行）
		log.Errorf("rate_limiter: redis get err: %v", err)
		return false
	}
	if countStr == "" {
		// 无记录，未限流
		return false
	}

	failCount, err := strconv.Atoi(countStr)
	if err != nil {
		log.Errorf("rate_limiter: redis get fail: %v", err)
		return false
	}

	if failCount < cfg.MaxAttempts {
		return false
	}

	// 已被限流，获取剩余冷却时间
	remainingSecs := int64(cfg.Cooldown.Seconds())
	ttl, ttlErr := redis.Sys().RedisTTL(ctx.Request.Context(), key)
	if ttlErr == nil && ttl.Seconds() > 0 {
		remainingSecs = int64(ttl.Seconds())
	}

	minutes := (remainingSecs + 59) / 60
	if minutes < 1 {
		minutes = 1
	}
	msg := gin_util.I18nKey(ctx, RateLimitedI18nKey, strconv.FormatInt(minutes, 10))
	gin_util.ResponseDetail(ctx, http.StatusBadRequest, codes.Code(err_code.Code_BFFRateLimited), nil, msg)
	ctx.Abort()
	return true
}

// handlePostCheck 后置处理：根据业务 handler 的响应结果决定是否计数
func handlePostCheck(ctx *gin.Context, cfg RateLimiterConfig, key string, bw *responseBufferWriter) {
	// 从 context 中读取结果判断业务是否成功（ResponseDetail 会将结果写入 gin_util.RESULT）
	resultStr := ctx.GetString(gin_util.RESULT)
	if resultStr == "" {
		// 没有结果，刷新原始响应
		ctx.Writer = bw.ResponseWriter
		bw.flushToReal()
		return
	}

	// 解析响应 JSON，提取 code 和原始错误信息
	var result struct {
		Code int64  `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		log.Errorf("rate_limiter: unmarshal result err: %v", err)
		// 解析失败，刷新原始响应
		ctx.Writer = bw.ResponseWriter
		bw.flushToReal()
		return
	}

	// 不在计数列表中
	if !isErrCode(cfg, result.Code) {
		// 成功（code == 0）且配置了 ResetOnSuccess 时，清除计数器
		if result.Code == int64(err_code.Code_OK) && cfg.ResetOnSuccess {
			_ = redis.Sys().Del(ctx.Request.Context(), key)
		}
		ctx.Writer = bw.ResponseWriter
		bw.flushToReal()
		return
	}

	// 匹配计数列表：原子递增计数并重置过期时间
	count, err := redis.Sys().IncrWithExpire(ctx.Request.Context(), key, cfg.Cooldown)
	if err != nil {
		// Redis 错误，fail-open，不修改响应
		log.Errorf("rate_limiter: redis incr err: %v", err)
		ctx.Writer = bw.ResponseWriter
		bw.flushToReal()
		return
	}

	// Code_OK 刷新原始响应
	if result.Code == int64(err_code.Code_OK) {
		ctx.Writer = bw.ResponseWriter
		bw.flushToReal()
		return
	}

	// 重写限流消息
	remaining := cfg.MaxAttempts - int(count)
	var msg string
	if remaining <= 0 {
		// 达到限流上限，获取剩余冷却时间
		remainingSecs := int64(cfg.Cooldown.Seconds())
		ttl, ttlErr := redis.Sys().RedisTTL(ctx.Request.Context(), key)
		if ttlErr == nil && ttl.Seconds() > 0 {
			remainingSecs = int64(ttl.Seconds())
		}
		minutes := (remainingSecs + 59) / 60
		if minutes < 1 {
			minutes = 1
		}
		msg = gin_util.I18nKey(ctx, RateLimitedI18nKey, strconv.FormatInt(minutes, 10))
	} else {
		// 还有剩余次数，追加提示
		msg = gin_util.I18nKey(ctx, FailedWithAttemptsI18nKey, result.Msg, strconv.Itoa(remaining))
	}

	// 丢弃缓冲的原始响应，重写为限流消息
	ctx.Writer = bw.ResponseWriter
	gin_util.ResponseDetail(ctx, http.StatusBadRequest, codes.Code(err_code.Code_BFFRateLimited), nil, msg)
}

// isErrCode 判断响应 code 是否需要计入限流
// ErrCodes 为空时，所有响应都计入（不管成功还是失败）
// ErrCodes 非空时，只有列表中的 code 才计入
func isErrCode(cfg RateLimiterConfig, code int64) bool {
	if len(cfg.ErrCodes) == 0 {
		return true
	}
	for _, c := range cfg.ErrCodes {
		if code == c {
			return true
		}
	}
	return false
}
