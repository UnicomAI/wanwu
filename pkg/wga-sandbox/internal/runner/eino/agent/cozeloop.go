package main

import (
	"log"
	"os"
	"sync"

	ccb "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"

	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/eino/agent/shared"
)

var (
	cozeloopHandler callbacks.Handler

	cozeloopInitOnce sync.Once
	cozeloopInitErr  error
)

// Init 初始化 CozeLoop 客户端并注册全局回调 handler。
// 多次调用安全：仅首次生效，后续调用直接返回首次结果。
// 环境变量：
//   - COZELOOP_ENABLED:      "true" 启用（缺省禁用，直接 no-op）
//   - COZELOOP_WORKSPACE_ID: workspace id
//   - COZELOOP_API_TOKEN:    access token
//   - COZELOOP_API_BASE_URL: 自部署场景的 API base URL（可选）
func Init() error {
	cozeloopInitOnce.Do(func() {
		cozeloopInitErr = setupCozeloop()
	})
	return cozeloopInitErr
}

func setupCozeloop() error {
	if os.Getenv("COZELOOP_ENABLED") != "true" {
		log.Println("[CozeLoop] disabled (set COZELOOP_ENABLED=true to enable)")
		return nil
	}

	workspaceID := os.Getenv("COZELOOP_WORKSPACE_ID")
	apiToken := os.Getenv("COZELOOP_API_TOKEN")
	if workspaceID == "" || apiToken == "" {
		log.Println("[CozeLoop] COZELOOP_WORKSPACE_ID or COZELOOP_API_TOKEN not set, skipping initialization")
		return nil
	}

	opts := []cozeloop.Option{
		cozeloop.WithWorkspaceID(workspaceID),
		cozeloop.WithAPIToken(apiToken),
	}
	if apiBaseURL := os.Getenv("COZELOOP_API_BASE_URL"); apiBaseURL != "" {
		opts = append(opts, cozeloop.WithAPIBaseURL(apiBaseURL))
		log.Printf("[CozeLoop] using custom API base URL: %s", shared.SanitizeForLog(apiBaseURL))
	}

	client, err := cozeloop.NewClient(opts...)
	if err != nil {
		log.Printf("[CozeLoop] create client failed: %v", err)
		return err
	}

	cozeloopHandler = ccb.NewLoopHandler(client)
	callbacks.AppendGlobalHandlers(cozeloopHandler)

	log.Println("[CozeLoop] initialized successfully")
	return nil
}
