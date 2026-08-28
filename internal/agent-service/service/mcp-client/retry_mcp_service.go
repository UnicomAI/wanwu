package mcp_client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

/**
 * 可以重试的mcp客户端，默认重试2次，每次重试间隔阶梯等待
 */

var defaultRetry = &Retry{
	Span:     time.Second * 3,
	StepSpan: true,
	Times:    2,
}

type Retry struct {
	Span     time.Duration //重试间隔
	StepSpan bool          // 是否递增间隔,第一次等待 span 第二次等待 span * 2 第三次 等待 span * 3
	Times    int           // 重试次数
}

type RetryMcpClient struct {
	client *client.Client
	retry  *Retry
	log    bool
}

func NewDefaultRetryMcpClient(client *client.Client) *RetryMcpClient {
	return NewRetryMcpClient(client, defaultRetry, true)
}

func NewRetryMcpClient(client *client.Client, retry *Retry, log bool) *RetryMcpClient {
	if retry == nil {
		retry = defaultRetry
	}
	return &RetryMcpClient{
		client: client,
		retry:  retry,
		log:    log,
	}
}

// Start initiates the connection to the server.
// Must be called before using the client.
func (r *RetryMcpClient) Start(ctx context.Context) error {
	return r.client.Start(ctx)
}

// Initialize sends the initial connection request to the server
func (r *RetryMcpClient) Initialize(
	ctx context.Context,
	request mcp.InitializeRequest,
) (*mcp.InitializeResult, error) {
	return r.client.Initialize(ctx, request)
}

// Ping checks if the server is alive
func (r *RetryMcpClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx)

}

// ListResourcesByPage manually list resources by page.
func (r *RetryMcpClient) ListResourcesByPage(
	ctx context.Context,
	request mcp.ListResourcesRequest,
) (*mcp.ListResourcesResult, error) {
	return r.client.ListResourcesByPage(ctx, request)
}

// ListResources requests a list of available resources from the server
func (r *RetryMcpClient) ListResources(
	ctx context.Context,
	request mcp.ListResourcesRequest,
) (*mcp.ListResourcesResult, error) {
	return r.client.ListResources(ctx, request)
}

// ListResourceTemplatesByPage manually list resource templates by page.
func (r *RetryMcpClient) ListResourceTemplatesByPage(
	ctx context.Context,
	request mcp.ListResourceTemplatesRequest,
) (*mcp.ListResourceTemplatesResult,
	error) {
	return r.client.ListResourceTemplatesByPage(ctx, request)
}

// ListResourceTemplates requests a list of available resource templates from the server
func (r *RetryMcpClient) ListResourceTemplates(
	ctx context.Context,
	request mcp.ListResourceTemplatesRequest,
) (*mcp.ListResourceTemplatesResult,
	error) {
	return r.client.ListResourceTemplates(ctx, request)
}

// ReadResource reads a specific resource from the server
func (r *RetryMcpClient) ReadResource(
	ctx context.Context,
	request mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {
	return r.client.ReadResource(ctx, request)
}

// Subscribe requests notifications for changes to a specific resource
func (r *RetryMcpClient) Subscribe(ctx context.Context, request mcp.SubscribeRequest) error {
	return r.client.Subscribe(ctx, request)
}

// Unsubscribe cancels notifications for a specific resource
func (r *RetryMcpClient) Unsubscribe(ctx context.Context, request mcp.UnsubscribeRequest) error {
	return r.client.Unsubscribe(ctx, request)
}

// ListPromptsByPage manually list prompts by page.
func (r *RetryMcpClient) ListPromptsByPage(
	ctx context.Context,
	request mcp.ListPromptsRequest,
) (*mcp.ListPromptsResult, error) {
	return r.client.ListPromptsByPage(ctx, request)
}

// ListPrompts requests a list of available prompts from the server
func (r *RetryMcpClient) ListPrompts(
	ctx context.Context,
	request mcp.ListPromptsRequest,
) (*mcp.ListPromptsResult, error) {
	return r.client.ListPrompts(ctx, request)
}

// GetPrompt retrieves a specific prompt from the server
func (r *RetryMcpClient) GetPrompt(
	ctx context.Context,
	request mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	return r.client.GetPrompt(ctx, request)
}

// ListToolsByPage manually list tools by page.
func (r *RetryMcpClient) ListToolsByPage(
	ctx context.Context,
	request mcp.ListToolsRequest,
) (*mcp.ListToolsResult, error) {
	return r.client.ListToolsByPage(ctx, request)

}

// ListTools requests a list of available tools from the server
func (r *RetryMcpClient) ListTools(
	ctx context.Context,
	request mcp.ListToolsRequest,
) (*mcp.ListToolsResult, error) {
	return r.client.ListTools(ctx, request)
}

// CallTool invokes a specific tool on the server
func (r *RetryMcpClient) CallTool(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	if r.retry != nil {
		for i := 0; i < r.retry.Times; i++ {
			result, err := callTool(r.client, ctx, request, r.log, i)
			if err == nil && !result.IsError {
				return result, nil
			}
			var duration = r.retry.Span
			if r.retry.StepSpan {
				duration = r.retry.Span * time.Duration(i+1)
			}
			time.Sleep(duration)
		}
	}
	result, err := callTool(r.client, ctx, request, r.log, 0)
	if err != nil {
		resp, err1 := response.ToolErrResp(err)
		if err1 != nil {
			return result, err
		}
		return &mcp.CallToolResult{
			IsError: false,
			Content: []mcp.Content{
				&mcp.TextContent{
					Type: "text",
					Text: resp,
				},
			},
		}, nil
	}
	return result, nil
}

// SetLevel sets the logging level for the server
func (r *RetryMcpClient) SetLevel(ctx context.Context, request mcp.SetLevelRequest) error {
	return r.client.SetLevel(ctx, request)

}

// Complete requests completion options for a given argument
func (r *RetryMcpClient) Complete(
	ctx context.Context,
	request mcp.CompleteRequest,
) (*mcp.CompleteResult, error) {
	return r.client.Complete(ctx, request)
}

// Close client connection and cleanup resources
func (r *RetryMcpClient) Close() error {
	return r.client.Close()
}

// OnNotification registers a handler for notifications
func (r *RetryMcpClient) OnNotification(handler func(notification mcp.JSONRPCNotification)) {
	r.client.OnNotification(handler)
}

func callTool(client client.MCPClient, ctx context.Context,
	request mcp.CallToolRequest, logFlag bool, retryTimes int) (*mcp.CallToolResult, error) {
	result, err := client.CallTool(ctx, request)
	if logFlag {
		defer util.PrintPanicStack()
		params, _ := json.Marshal(request)
		if err != nil {
			log.Errorf("CallTool %s err: %v, retryTimes: %d", string(params), err, retryTimes)
		} else {
			var message = "success"
			if result.IsError {
				message = "failed"
			}
			marshalJSON, _ := result.MarshalJSON()
			log.Infof("CallTool %s %s: %v, retryTimes: %d", string(params), message, string(marshalJSON), retryTimes)
		}
	}
	return result, err
}
