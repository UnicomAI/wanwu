package grpc

import (
	"context"
	"net"
	"time"

	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/server/grpc/rag"
	"google.golang.org/grpc"

	"github.com/UnicomAI/wanwu/internal/rag-service/client"
	"github.com/UnicomAI/wanwu/internal/rag-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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
	// 初始化微服务
	if err := rag.StartService(); err != nil {
		log.Fatalf("init service err: %v", err)
	}
	log.Infof("init service success")

	s.serv = trace_util.NewGrpcTracerServer(
		[]grpc.UnaryServerInterceptor{trace_util.LoggingUnaryGRPC()},
		[]grpc.StreamServerInterceptor{trace_util.LoggingStreamGRPC()},
	)

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(s.serv, healthcheck)
	// 注册operate_service
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
