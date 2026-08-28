package grpc_provider

import (
	"context"
	"net"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	knowledge_log "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/knowledge-log"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/tracer"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var grpcProvider = GrpcProvider{}
var serverInterceptor = ServerInterceptor{}

type GrpcProvider struct {
	Serv *grpc.Server
}

func init() {
	pkg.AddContainer(grpcProvider)
}

// ServerInterceptor 拦截器结构体
type ServerInterceptor struct{}

// Unary 拦截服务方法的处理
func (si *ServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now().UnixMilli()
		var err error
		var resp interface{}
		defer util.PrintPanicStackWithCall(func(panicOccur bool, err2 error) {
			if panicOccur {
				err = err2
			}
			knowledge_log.LogAccessPB(ctx, "grpc", info.FullMethod, req, resp, err, start)
		})
		// 调用原始的处理函数
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func (c GrpcProvider) LoadType() string {
	return "grpc-provider"
}

func (c GrpcProvider) Load() error {
	configInfo := config.GetConfig()

	c.Serv = trace_util.NewGrpcTracerServer([]grpc.UnaryServerInterceptor{serverInterceptor.Unary()}, nil)

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(c.Serv, healthcheck)

	// register service
	err := RegisterAllGrpcService(c.Serv)
	if err != nil {
		return err
	}

	// listen
	lis, err := net.Listen("tcp", configInfo.Server.GrpcEndpoint)
	if err != nil {
		return err
	}

	// serve
	go func() {
		err = c.Serv.Serve(lis)
		if err != nil {
			log.Fatalf("knowledge grpc server.Serve() failed, err: %v", err)
		} else {
			log.Infof("knowledge grpc server.Serve() succeed.")
		}
	}()

	log.Infof("start knowledge grpc server at: %s", configInfo.Server.GrpcEndpoint)
	return nil
}

func (c GrpcProvider) StopPriority() int {
	return pkg.GrpcPriority
}

func (c GrpcProvider) Stop() error {
	log.Infof("knowledge closing grpc server...")
	if c.Serv == nil {
		log.Infof("knowledge closing grpc server is nil...")
		return nil
	}
	stopped := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("knowledge recover grpc server panic: %v", r)
			}
			close(stopped)
		}()
		c.Serv.GracefulStop()
		log.Infof("knowledge close grpc server gracefully")
	}()

	t := time.NewTimer(time.Minute)
	select {
	case <-t.C:
		c.Serv.Stop()
		log.Errorf("knowledge close grpc server forced")
	case <-stopped:
		t.Stop()
	}
	return nil
}
