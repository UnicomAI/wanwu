package task

import (
	"errors"
	"fmt"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/app-service/config"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"google.golang.org/grpc"
)

// 下游微服务 gRPC client 与 orm client 的集中初始化，由 app-service main 调一次 Init 完成注入。

var (
	// assistant assistant-service gRPC client，对话日志导出查询对话详情用。
	assistant assistant_service.AssistantServiceClient
)

// ormClient 对话日志导出 worker 访问 DB 的 orm client。
var ormClient *orm.Client

// Init 注入 orm client 并建立 assistant-service gRPC 连接，由 app-service main 在 client 初始化后、InitAsync 前调用。
func Init(c *orm.Client) error {
	if c == nil {
		return errors.New("orm client is nil")
	}
	ormClient = c

	assistantConn, err := newConn(config.Cfg().Assistant.Host)
	if err != nil {
		return fmt.Errorf("init assistant-service connection err: %v", err)
	}
	assistant = assistant_service.NewAssistantServiceClient(assistantConn)
	return nil
}

func newConn(host string) (*grpc.ClientConn, error) {
	return trace_util.NewGrpcTracerConn(host, nil)
}
