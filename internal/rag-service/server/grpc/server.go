package grpc

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"

	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/server/grpc/rag"
	"google.golang.org/grpc"

	"github.com/UnicomAI/wanwu/internal/rag-service/client"
	"github.com/UnicomAI/wanwu/internal/rag-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	cfg  *config.Config
	serv *grpc.Server
	rag  *rag.Service
}

func NewServer(cfg *config.Config, cli client.IClient) (*Server, error) {
	s := &Server{
		cfg: cfg,
		rag: rag.NewService(cli),
	}
	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	if s.serv != nil {
		return nil
	}
	// Initialize microservices
	if err := rag.StartService(); err != nil {
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
		grpc.ChainUnaryInterceptor(grpc_recovery.UnaryServerInterceptor(opts...)),
		grpc.ChainStreamInterceptor(grpc_recovery.StreamServerInterceptor(opts...)),
	}
	s.serv = grpc.NewServer(serverOptions...)

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(s.serv, healthcheck)
	// Register operate_service
	rag_service.RegisterRagServiceServer(s.serv, s.rag)
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

func (s *Server) Stop(ctx context.Context) {
	if s.serv == nil {
		return
	}

	log.Infof("closing grpc server...")
	stopped := make(chan struct{})
	go func() {
		s.serv.GracefulStop()
		log.Infof("close grpc server gracefully")
		close(stopped)
	}()

	cancelCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	select {
	case <-cancelCtx.Done():
		s.serv.Stop()
		log.Errorf("close grpc server forced")
	case <-stopped:
	}
}
