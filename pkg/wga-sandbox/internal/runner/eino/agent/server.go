package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	chatmodel "github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/eino/agent/chat-model"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/eino/agent/shared"
	"github.com/cloudwego/eino/adk"
	"github.com/joho/godotenv"
)

const (
	maxRequestBodyBytes = 10 << 20 // 10MB
	requestTimeout      = 30 * time.Minute
	defaultAgentType    = "chat-model"
)

// agentBuilders 是 agent_type → 构造函数的注册表。
// 新增 agent 类型只需在此追加一项，handler 无需改动。
var agentBuilders = map[string]func(ctx context.Context, cfg shared.AppConfig) (shared.AgentApp, error){
	"chat-model": chatmodel.NewApp,
}

type httpServer struct{}

type chatRequest struct {
	Messages []adk.Message `json:"messages"`
}

func newHTTPServer() *httpServer {
	return &httpServer{}
}

func (s *httpServer) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/chat", s.handleChat)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func (s *httpServer) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workspace := r.URL.Query().Get("workspace")
	if workspace == "" {
		http.Error(w, "workspace parameter is required", http.StatusBadRequest)
		return
	}
	agentTypeQuery := r.URL.Query().Get("agent_type")

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Chat] decode request body failed: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Messages) == 0 {
		log.Printf("[Chat] empty messages received")
		http.Error(w, "messages is required", http.StatusBadRequest)
		return
	}

	log.Printf("[Chat] workspace=%s messages=%d", shared.SanitizeForLog(workspace), len(req.Messages))

	// SSE header 必须在任何业务调用之前写出，确保后续无论 NewApp / Query / ProcessEvents
	// 哪一步失败，都能用 SSE data 行回写一条 assistant+stop 兜底消息，
	// 而不是退化为 http.Error(500) 让上游 converter 收到非 schema.Message 内容而丢弃。
	sseWriter, ok := beginSSEResponse(w)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	sessionID := "session-" + uuid.New().String()
	log.Printf("[Chat] start sessionID=%s", sessionID)

	cfg, agentType := resolveAgent(workspace, agentTypeQuery)

	// sessionCtx：可 cancel-with-cause 的会话上下文。
	// 用途：sandbox 侧连续 [BLOCKED:...] 到阈值触发 HaltState.haltFn → cancelSession(err) →
	// eino runner iterator 关闭 → ProcessEvents 自然收尾 → 下面 defer 通过 context.Cause
	// 拿到熔断原因，作为 finalErr 走 BuildFinalAgentEvent(FinalErrorSourceAgent, err)
	// 路径下发 assistant+stop 消息 content = "error[agent]: consecutive security blocks reached threshold N"。
	sessionCtx, cancelSession := context.WithCancelCause(ctx)
	defer cancelSession(nil)

	cfg.Halt = shared.NewHaltState(sandboxHaltThreshold, func(err error) {
		log.Printf("[Chat] halt triggered sessionID=%s: %v", sessionID, err)
		cancelSession(err)
	})

	// 兜底保证：handler 退出前必发一条 assistant+stop 消息。
	// runChat 在自然 / 错误收尾时回写 sentFinal=true。
	sentFinal := false
	var finalErr error

	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("[Chat] panic recovered sessionID=%s panic=%v", sessionID, rec)
			if !sentFinal {
				sseWriter.WriteAgentEvent(shared.BuildFinalAgentEvent(
					shared.FinalErrorSourceAgent,
					fmt.Errorf("panic: %v", rec),
				))
			}
			return
		}
		if !sentFinal {
			err := finalErr
			if err == nil {
				// 优先使用 sessionCtx 的 cancel cause（熔断原因等有信息量的 error）；
				// 若 cause 只是普通 context.Canceled / DeadlineExceeded，退回到父 ctx.Err()。
				if cause := context.Cause(sessionCtx); cause != nil && cause != context.Canceled && cause != context.DeadlineExceeded {
					err = cause
				} else if ctx.Err() != nil {
					err = ctx.Err()
				}
			}
			sseWriter.WriteAgentEvent(shared.BuildFinalAgentEvent(shared.FinalErrorSourceAgent, err))
		}
	}()

	sentFinal, finalErr = runChat(sessionCtx, cfg, agentType, req.Messages, sseWriter)
	log.Printf("[Chat] done sessionID=%s sentFinal=%v err=%v", sessionID, sentFinal, finalErr)
}

// sandboxHaltThreshold sandbox 会话内连续 [BLOCKED:...] 触发强制熔断的阈值。
// 阈值取 3 的理由：合法业务下第一次 precheck 误伤后 LLM 通常一两次内切到正确路径；
// 连续 3 次以上 BLOCKED 更明确指示"agent 在反复试图绕过安全规则"，误伤概率极低。
const sandboxHaltThreshold = 3

// beginSSEResponse 校验 ResponseWriter 是否支持 Flush，并写出 SSE 响应头。
func beginSSEResponse(w http.ResponseWriter) (shared.SSEWriter, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("[Chat] streaming not supported")
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	return shared.NewHTTPSSEWriter(w, flusher), true
}

// resolveAgent 解析 agent_type 与 AppConfig：
// agent_type 优先级 query → workspace/.env 的 EINO_AGENT_TYPE → "chat-model"。
func resolveAgent(workspace, agentTypeQuery string) (shared.AppConfig, string) {
	envMap := loadEnvFromFile(filepath.Join(workspace, ".env"))

	agentType := agentTypeQuery
	if agentType == "" {
		agentType = getEnvValue(envMap, "EINO_AGENT_TYPE")
	}
	if agentType == "" {
		agentType = defaultAgentType
	}
	log.Printf("[Chat] agent_type=%s", shared.SanitizeForLog(agentType))

	cfg := shared.AppConfig{
		Workspace: workspace,
		APIKey:    getEnvValue(envMap, "OPENAI_API_KEY"),
		BaseURL:   getEnvValue(envMap, "OPENAI_BASE_URL"),
		ModelID:   getEnvValue(envMap, "OPENAI_MODEL_ID"),
	}
	if raw := getEnvValue(envMap, "OPENAI_LLM_PARAMS"); raw != "" {
		var p shared.LLMParams
		if err := json.Unmarshal([]byte(raw), &p); err == nil {
			cfg.Params = &p
		} else {
			log.Printf("[Chat] unmarshal OPENAI_LLM_PARAMS failed: %v", err)
		}
	}
	return cfg, agentType
}

// runChat 构造 agent 应用并消费事件流。
// 返回 sentFinal 由 ProcessEvents 决定；返回的 err 供 handler 兜底使用。
func runChat(ctx context.Context, cfg shared.AppConfig, agentType string,
	messages []adk.Message, sseWriter shared.SSEWriter) (sentFinal bool, err error) {

	builder, ok := agentBuilders[agentType]
	if !ok {
		return false, fmt.Errorf("invalid agent_type: %s", agentType)
	}

	app, err := builder(context.Background(), cfg)
	if err != nil {
		log.Printf("[Chat] create app failed: %v", err)
		return false, fmt.Errorf("failed to create app: %w", err)
	}
	defer func() {
		if cerr := app.Close(); cerr != nil {
			log.Printf("[Chat] close app failed: %v", cerr)
		}
	}()

	iter := app.Query(ctx, messages)
	if iter == nil {
		log.Printf("[Chat] query returned nil iterator")
		return false, fmt.Errorf("query returned nil iterator")
	}

	_, sentFinal = shared.ProcessEvents(ctx, iter, sseWriter)
	return sentFinal, nil
}

func loadEnvFromFile(path string) map[string]string {
	envMap, err := godotenv.Read(path)
	if err != nil {
		log.Printf("[Env] read .env from %s failed: %v", shared.SanitizeForLog(path), err)
		return map[string]string{}
	}
	return envMap
}

func getEnvValue(envMap map[string]string, key string) string {
	if val, ok := envMap[key]; ok {
		return val
	}
	return os.Getenv(key)
}
