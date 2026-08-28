package agent_tool

import (
	"context"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	mcp_client "github.com/UnicomAI/wanwu/internal/agent-service/service/mcp-client"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	mcp_util "github.com/UnicomAI/wanwu/pkg/mcp-util"
	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcpTypes "github.com/mark3labs/mcp-go/mcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type MCPServerInfo struct {
	Transport    string   `json:"transport"`
	URL          string   `json:"url"`
	ToolNameList []string `json:"toolNameList"`
}

// createMCPClient 根据 transport 类型创建 MCP 客户端
// transport: "sse" 或 "streamable"
func createMCPClient(ctx context.Context, mcpToolInfo *request.MCPToolInfo) (client.MCPClient, error) {
	var mcpClient *client.Client
	var transportType = mcpToolInfo.Transport
	url, headers, err := mcp_util.MergeMcpParams(mcpToolInfo.URL, mcpToolInfo.ApiAuth, mcpToolInfo.Headers)
	if err != nil {
		log.Errorf("failed to merge mcp params: %v", err)
		return nil, fmt.Errorf("failed to merge mcp params: %w", err)
	}
	if headers == nil {
		headers = make(map[string]string)
	}
	//header 注入traceID
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))

	switch transportType {
	case constant.MCPTransportStreamable:
		// 创建 StreamableHTTP 客户端
		if len(headers) > 0 {
			mcpClient, err = client.NewStreamableHttpClient(url, transport.WithHTTPHeaders(headers))
		} else {
			mcpClient, err = client.NewStreamableHttpClient(url)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create StreamableHTTP MCP client: %w", err)
		}
	case constant.MCPTransportSSE:
		// 默认使用 SSE 客户端
		if len(headers) > 0 {
			mcpClient, err = client.NewSSEMCPClient(url, transport.WithHeaders(headers))
		} else {
			mcpClient, err = client.NewSSEMCPClient(url)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create SSE MCP client: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", transportType)
	}

	retryMcpClient := mcp_client.NewDefaultRetryMcpClient(mcpClient)

	// 启动客户端
	err = retryMcpClient.Start(ctx)
	if err != nil {
		_ = retryMcpClient.Close()
		return nil, fmt.Errorf("failed to start MCP client: %w", err)
	}

	// 初始化 MCP 客户端
	initCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	initRequest := mcpTypes.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcpTypes.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcpTypes.Implementation{
		Name:    "eino-mcp-client",
		Version: "0.1.0",
	}
	initRequest.Params.Capabilities = mcpTypes.ClientCapabilities{}

	_, err = retryMcpClient.Initialize(initCtx, initRequest)
	if err != nil {
		_ = retryMcpClient.Close()
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	log.Infof("MCP client (%s) initialized successfully", transportType)
	return retryMcpClient, nil
}

func GetToolsFromMCPServers(ctx context.Context, toolParamsList []*request.MCPToolInfo) ([]tool.BaseTool, map[string]*request.ToolConfig, error) {
	if len(toolParamsList) == 0 {
		return nil, nil, nil
	}

	var allTools []tool.BaseTool
	var toolMap = make(map[string]*request.ToolConfig)

	for _, serverInfo := range toolParamsList {
		log.Infof("Connecting to MCP server: %v", serverInfo)

		mcpClient, err := createMCPClient(ctx, serverInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create MCP client for %v: %v", serverInfo, err)
		}
		// 注意:不要在这里关闭客户端,因为工具在后续使用时还需要这个连接
		// defer mcpClient.Close()

		tools, err := mcp.GetTools(ctx, &mcp.Config{
			Cli:          mcpClient,
			ToolNameList: serverInfo.ToolNameList,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get mcp tools from %v: %v", serverInfo, err)
		}

		log.Infof("Loaded %d tools from %v", len(tools), serverInfo)
		if len(serverInfo.ToolNameList) > 0 {
			//mcp 的方法名先不做替换，因为mcp的函数名基本都是符合规则一般不会有特殊字符
			for _, toolName := range serverInfo.ToolNameList {
				toolMap[toolName] = &request.ToolConfig{
					Avatar:   serverInfo.Avatar,
					ToolName: toolName,
					ToolID:   toolName,
				}
			}
		}
		allTools = append(allTools, tools...)
	}

	return allTools, toolMap, nil
}
