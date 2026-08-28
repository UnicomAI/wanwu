package mcp

import (
	"fmt"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"google.golang.org/grpc"
)

var (
	app app_service.AppServiceClient
)

func StartService() error {
	// grpc connections
	AppConn, err := newConn(config.Cfg().App.Host)
	if err != nil {
		return fmt.Errorf("init app-service connection err: %v", err)
	}
	app = app_service.NewAppServiceClient(AppConn)
	log.Infof("App init success")
	log.Infof("App: %s", config.Cfg().App.Host)
	return nil
}

func newConn(host string) (*grpc.ClientConn, error) {
	return trace_util.NewGrpcTracerConn(host, nil)
}
