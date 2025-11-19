package grpc

import (
	"fmt"
	"net"
	"runtime/debug"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/internal/assistant-service/server/grpc/assistant"
	"github.com/UnicomAI/wanwu/internal/assistant-service/server/grpc/interceptor"
	"github.com/UnicomAI/wanwu/pkg/log"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	cfg  *config.Config
	serv *grpc.Server

	assistant *assistant.Service
}

func NewServer(cfg *config.Config, cli client.IClient) (*Server, error) {
	s := &Server{
		cfg:       cfg,
		assistant: assistant.NewService(cli),
	}
	return s, nil
}

func (s *Server) Start() error {
	if s.serv != nil {
		return nil
	}
	// 初始化微服务 [EN] Initialize microservices
	if err := assistant.StartService(); err != nil {
		log.Fatalf("init service err: %v", err)
	}
	log.Infof("init service success")

	// init
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Errorf("[PANIC] %v\n%v", p, string(debug.Stack()))
			return status.Error(codes.Internal, fmt.Sprintf("panic: %v", p))
		}),
	}
	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(s.cfg.Server.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(s.cfg.Server.MaxRecvMsgSize),
		grpc.ChainUnaryInterceptor(grpc_recovery.UnaryServerInterceptor(opts...), interceptor.LoggingUnaryGRPC()),
		grpc.ChainStreamInterceptor(grpc_recovery.StreamServerInterceptor(opts...)),
	}
	s.serv = grpc.NewServer(serverOptions...)

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(s.serv, healthcheck)

	// register service
	assistant_service.RegisterAssistantServiceServer(s.serv, s.assistant)

	// listen
	lis, err := net.Listen("tcp", s.cfg.Server.GrpcEndpoint)
	if err != nil {
		return err
	}

	// serve
	go func() {
		err = s.serv.Serve(lis)
		if err != nil {
			log.Fatalf("grpc server.Serve() failed, err: %v", err)
		}
	}()

	log.Infof("start grpc server at: %s", s.cfg.Server.GrpcEndpoint)
	return nil
}

func (s *Server) Stop() {
	if s.serv == nil {
		return
	}

	log.Infof("closing grpc server...")
	stopped := make(chan struct{})
	go func() {
		s.serv.GracefulStop()
		close(stopped)
		log.Infof("close grpc server gracefully")
	}()

	t := time.NewTimer(time.Minute)
	select {
	case <-t.C:
		s.serv.Stop()
		log.Errorf("close grpc server forced")
	case <-stopped:
		t.Stop()
	}
}
