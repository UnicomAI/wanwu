package grpc_provider

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	knowledge_log "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/knowledge-log"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var grpcProvider = GrpcProvider{}
var serverInterceptor = ServerInterceptor{}

type GrpcProvider struct {
	Serv *grpc.Server
}

func init() {
	pkg.AddContainer(grpcProvider)
}

// ServerInterceptor interceptor structure
type ServerInterceptor struct{}

// Unary interception service method processing
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
		// Call the original handler function
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func (c GrpcProvider) LoadType() string {
	return "grpc-provider"
}

func (c GrpcProvider) Load() error {
	configInfo := config.GetConfig()

	// init
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Errorf("[PANIC] %v\n%v", p, string(debug.Stack()))
			return status.Error(codes.Internal, fmt.Sprintf("panic: %v", p))
		}),
	}
	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(configInfo.Server.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(configInfo.Server.MaxRecvMsgSize),
		grpc.ChainUnaryInterceptor(grpc_recovery.UnaryServerInterceptor(opts...)),
		grpc.ChainStreamInterceptor(grpc_recovery.StreamServerInterceptor(opts...)),
		grpc.ChainUnaryInterceptor(serverInterceptor.Unary()),
	}
	c.Serv = grpc.NewServer(serverOptions...)

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
