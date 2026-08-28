package grpc

import (
	"context"
	"net"
	"time"

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/app-service/client"
	"github.com/UnicomAI/wanwu/internal/app-service/config"
	"github.com/UnicomAI/wanwu/internal/app-service/server/grpc/app"
	"github.com/UnicomAI/wanwu/internal/app-service/server/grpc/safety"
	"github.com/UnicomAI/wanwu/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
	cfg  *config.Config
	serv *grpc.Server

	app    *app.Service
	safety *safety.Service
}

func NewServer(cfg *config.Config, cli client.IClient) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		app:    app.NewService(cli),
		safety: safety.NewService(cli),
	}
	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	if s.serv != nil {
		return nil
	}

	// init
	s.serv = trace_util.NewGrpcTracerServer([]grpc.UnaryServerInterceptor{trace_util.LoggingUnaryGRPC()}, []grpc.StreamServerInterceptor{trace_util.LoggingStreamGRPC()})

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(s.serv, healthcheck)

	// register service
	app_service.RegisterAppServiceServer(s.serv, s.app)
	safety_service.RegisterSafetyServiceServer(s.serv, s.safety)

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
