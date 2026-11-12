package mcp_util

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"github.com/UnicomAI/wanwu/api/proto/common"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	mcp_util "github.com/UnicomAI/wanwu/pkg/mcp-util"
	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// headerTransport 是一个 http.RoundTripper 包装器，用于注入自定义请求头
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}
	otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
	return t.base.RoundTrip(req)
}

// newHTTPClientWithHeaders 创建带有header息的 HTTP 客户端
func newHTTPClientWithHeaders(headers map[string]string, strict bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	clientWithAuth := &http.Client{
		Transport: &headerTransport{
			base:    transport,
			headers: headers,
		},
	}
	if strict {
		transport.DialContext = url_util.SafeDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		})
		clientWithAuth.CheckRedirect = url_util.CheckRedirectPublic
	}
	return clientWithAuth
}

// ListToolsWithAuth 根据transport类型获取MCP工具列表，支持鉴权和自定义请求头。
// strict=true 时只允许访问公网地址，用于用户自定义 MCP。
func ListToolsWithAuth(ctx context.Context, url string, transportType string, auth *util.ApiAuthWebRequest, headers map[string]string, strict bool) ([]*protocol.Tool, error) {
	var transportClient transport.ClientTransport
	var err error
	if transportType == "" || url == "" {
		return nil, fmt.Errorf("transport type or url is empty")
	}

	mergedUrl, mergedHeaders, err := mcp_util.MergeMcpParams(url, buildApiAuth(auth), headers)
	if err != nil {
		return nil, err
	}

	// 构建带鉴权的 HTTP 客户端
	clientWithAuth := newHTTPClientWithHeaders(mergedHeaders, strict)

	switch transportType {
	case constant.MCPTransportStreamable:
		// 创建 StreamableHTTP 传输客户端
		transportClient, err = transport.NewStreamableHTTPClientTransport(mergedUrl,
			transport.WithStreamableHTTPClientOptionLogger(log.Log()),
			transport.WithStreamableHTTPClientOptionHTTPClient(clientWithAuth),
		)
		if err != nil {
			return nil, fmt.Errorf("mcp list tools (%v) init streamable transport err: %v", url, err)
		}
	case constant.MCPTransportSSE:
		// 默认使用 SSE 传输客户端
		transportClient, err = transport.NewSSEClientTransport(mergedUrl,
			transport.WithSSEClientOptionReceiveTimeout(time.Minute*2),
			transport.WithSSEClientOptionLogger(log.Log()),
			transport.WithSSEClientOptionHTTPClient(clientWithAuth),
		)
		if err != nil {
			return nil, fmt.Errorf("mcp list tools (%v) init sse transport err: %v", url, err)
		}
	default:
		return nil, fmt.Errorf("mcp list tools (%v) init transport err: %v", url, err)
	}

	// 初始化 MCP 客户端
	mcpClient, err := client.NewClient(transportClient)
	if err != nil {
		return nil, fmt.Errorf("mcp list tools (%v) init client err: %v", url, err)
	}
	defer func() { _ = mcpClient.Close() }()

	// 获取可用工具列表
	resp, err := mcpClient.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("mcp list tools (%v) err: %v", url, err)
	}
	return resp.Tools, nil
}

func buildApiAuth(auth *util.ApiAuthWebRequest) *common.ApiAuthWebRequest {
	if auth == nil {
		return nil
	}
	return &common.ApiAuthWebRequest{
		AuthType:           auth.AuthType,
		ApiKeyQueryParam:   auth.ApiKeyQueryParam,
		ApiKeyHeader:       auth.ApiKeyHeader,
		ApiKeyValue:        auth.ApiKeyValue,
		ApiKeyHeaderPrefix: auth.ApiKeyHeaderPrefix,
	}
}
