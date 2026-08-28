package grpc_consumer

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/pkg"
	agent_log "github.com/UnicomAI/wanwu/internal/agent-service/pkg/agent-log"
	_ "github.com/UnicomAI/wanwu/internal/agent-service/pkg/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var grpcConsumer = GrpcConsumer{}

type GrpcConsumer struct {
}

func init() {
	pkg.AddContainer(grpcConsumer)
}

func (c GrpcConsumer) LoadType() string {
	return "grpc-consumer"
}

func (c GrpcConsumer) Load() error {
	err := RegisterAllGrpcConsumerService()
	if err != nil {
		return fmt.Errorf("init grpc connection err: %v", err)
	}
	return nil
}

func (c GrpcConsumer) StopPriority() int {
	return pkg.GrpcPriority
}

func (c GrpcConsumer) Stop() error {
	return nil
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 预处理(pre-processing)
		start := time.Now().UnixMilli()
		// 获取正在运行程序的操作系统
		cos := runtime.GOOS
		// 将操作系统信息附加到传出请求

		ctx = metadata.AppendToOutgoingContext(ctx, "client-os", cos)

		// 可以看做是当前 RPC 方法，一般在拦截器中调用 invoker 能达到调用 RPC 方法的效果，当然底层也是 gRPC 在处理。
		// 调用RPC方法(invoking RPC method)
		err := invoker(ctx, method, req, reply, cc, opts...)

		// 后处理(post-processing) )
		agent_log.LogAccessPB(ctx, "grpc-consumer", method, req, reply, err, start)
		return err
	}
}
