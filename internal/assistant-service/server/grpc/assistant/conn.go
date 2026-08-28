package assistant

import (
	"fmt"

	knowledgeBase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"google.golang.org/grpc"
)

var (
	Knowledge knowledgeBase_service.KnowledgeBaseServiceClient
	MCP       mcp_service.MCPServiceClient
)

func StartService() error {
	// grpc connections
	knowledgeConn, err := newConn(config.Cfg().Knowledge.Host)
	if err != nil {
		return fmt.Errorf("init knowledgebase-service connection err: %v", err)
	}
	Knowledge = knowledgeBase_service.NewKnowledgeBaseServiceClient(knowledgeConn)
	log.Infof("Knowledge init success")
	log.Infof("Knowledge: %s", config.Cfg().Knowledge.Host)

	MCPConn, err := newConn(config.Cfg().MCP.Host)
	if err != nil {
		return fmt.Errorf("init mcp-service connection err: %v", err)
	}
	MCP = mcp_service.NewMCPServiceClient(MCPConn)
	log.Infof("MCP init success")
	log.Infof("MCP: %s", config.Cfg().MCP.Host)
	return nil
}

func newConn(host string) (*grpc.ClientConn, error) {
	return trace_util.NewGrpcTracerConn(host, nil)
}
