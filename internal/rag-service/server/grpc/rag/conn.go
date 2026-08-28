package rag

import (
	"fmt"

	knowledgeBase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"google.golang.org/grpc"
)

var (
	Knowledge knowledgeBase_service.KnowledgeBaseServiceClient
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
	return nil
}

func newConn(host string) (*grpc.ClientConn, error) {
	return trace_util.NewGrpcTracerConn(host, nil)
}
