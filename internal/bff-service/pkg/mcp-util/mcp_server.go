package mcp_util

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

var msMgr *mcpServerMgr

type mcpServer struct {
	sseServer           *server.Server
	sseHandler          *transport.SSEHandler
	sseTransport        transport.ServerTransport
	streamableServer    *server.Server
	streamableHandler   *transport.StreamableHTTPHandler
	streamableTransport transport.ServerTransport
}

type mcpServerMgr struct {
	mcpServers map[string]*mcpServer
	mu         sync.RWMutex
}

func Init(ctx context.Context) error {
	if msMgr != nil {
		return fmt.Errorf("mcp server already init")
	}
	msMgr = &mcpServerMgr{
		mcpServers: make(map[string]*mcpServer),
	}
	return nil
}

func CheckMCPServerExist(mcpServerId string) bool {
	msMgr.mu.RLock()
	defer msMgr.mu.RUnlock()
	_, exist := msMgr.mcpServers[mcpServerId]
	return exist
}

// --- mcp server ---

func StartMCPServer(ctx context.Context, mcpServerId string) error {
	if msMgr == nil {
		return fmt.Errorf("mcp server manager not init")
	}
	msMgr.mu.Lock()
	defer msMgr.mu.Unlock()
	if _, ok := msMgr.mcpServers[mcpServerId]; ok {
		return fmt.Errorf("mcp server(%v) already exist", mcpServerId)
	}

	sseTransport, sseHandler, err := transport.NewSSEServerTransportAndHandler("message",
		transport.WithSSEServerTransportAndHandlerOptionCopyParamKeys([]string{"key"}))
	if err != nil {
		return fmt.Errorf("new sse transport and hander with error: %v", err)
	}
	sseSrv, err := server.NewServer(sseTransport)
	if err != nil {
		return fmt.Errorf("new server with error: %v", err)
	}
	streamTransport, streamHandler, err := transport.NewStreamableHTTPServerTransportAndHandler(
		transport.WithStreamableHTTPServerTransportAndHandlerOptionStateMode(transport.Stateful))
	if err != nil {
		return fmt.Errorf("new sse transport and hander with error: %v", err)
	}
	streamSrv, err := server.NewServer(streamTransport)
	if err != nil {
		return fmt.Errorf("new server with error: %v", err)
	}

	msMgr.mcpServers[mcpServerId] = &mcpServer{
		sseServer:           sseSrv,
		sseHandler:          sseHandler,
		sseTransport:        sseTransport,
		streamableServer:    streamSrv,
		streamableHandler:   streamHandler,
		streamableTransport: streamTransport,
	}
	return nil
}

func ShutDownMCPServer(ctx context.Context, mcpServerId string) error {
	if msMgr == nil {
		return fmt.Errorf("mcp server manager not init")
	}
	msMgr.mu.Lock()
	defer msMgr.mu.Unlock()
	server, ok := msMgr.mcpServers[mcpServerId]
	if !ok {
		return fmt.Errorf("mcp server(%v) not exist", mcpServerId)
	}

	err := server.sseServer.Shutdown(ctx)
	if err != nil {
		return err
	}
	err = server.streamableServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	delete(msMgr.mcpServers, mcpServerId)
	return nil
}

// --- mcp server tool ---

func RegisterMCPServerTools(mcpServerId string, tools []*McpTool) error {
	if msMgr == nil {
		return fmt.Errorf("mcp server manager not init")
	}
	msMgr.mu.Lock()
	defer msMgr.mu.Unlock()
	server, ok := msMgr.mcpServers[mcpServerId]
	if !ok {
		return fmt.Errorf("mcp server(%v) not exist", mcpServerId)
	}

	for _, tool := range tools {
		server.sseServer.RegisterTool(tool.tool, tool.handler)
		server.streamableServer.RegisterTool(tool.tool, tool.handler)
	}
	return nil
}

func UnRegisterMCPServerTools(mcpServerId string, tools []string) error {
	if msMgr == nil {
		return fmt.Errorf("mcp server manager not init")
	}
	msMgr.mu.Lock()
	defer msMgr.mu.Unlock()
	server, ok := msMgr.mcpServers[mcpServerId]
	if !ok {
		return fmt.Errorf("mcp server(%v) not exist", mcpServerId)
	}

	for _, tool := range tools {
		server.sseServer.UnregisterTool(tool)
		server.streamableServer.UnregisterTool(tool)
	}
	return nil
}

// --- mcp server handler ---

func HandleSSE(mcpServerId string, resp http.ResponseWriter, req *http.Request) error {
	server, ok := getMcpServer(mcpServerId)
	if !ok {
		return fmt.Errorf("mcp server(%v) not exist", mcpServerId)
	}
	server.sseHandler.HandleSSE().ServeHTTP(resp, req)
	return nil
}

func HandleMessage(mcpServerId string, resp http.ResponseWriter, req *http.Request) error {
	server, ok := getMcpServer(mcpServerId)
	if !ok {
		return fmt.Errorf("mcp server(%v) not exist", mcpServerId)
	}
	server.sseHandler.HandleMessage().ServeHTTP(resp, req)
	return nil
}

func HandleStreamable(mcpServerId string, resp http.ResponseWriter, req *http.Request) error {
	server, ok := getMcpServer(mcpServerId)
	if !ok {
		return fmt.Errorf("mcp server(%v) not exist", mcpServerId)
	}
	server.streamableHandler.HandleMCP().ServeHTTP(resp, req)
	return nil
}

// --- internal ---

func getMcpServer(mcpServerId string) (*mcpServer, bool) {
	msMgr.mu.RLock()
	defer msMgr.mu.RUnlock()
	server, ok := msMgr.mcpServers[mcpServerId]
	return server, ok
}
