package grpc

import (
	"context"
	"net"
	"time"

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"

	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/operate-service/client"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/operate-service/config"
	"github.com/UnicomAI/wanwu/internal/operate-service/server/grpc/notice"
	"github.com/UnicomAI/wanwu/internal/operate-service/server/grpc/operate"
	"github.com/UnicomAI/wanwu/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
	cfg  *config.Config
	serv *grpc.Server

	operate *operate.Service
	// notice 消息中心（独立 service，与 OperateService 并列双注册）
	notice *notice.Service
}

func NewServer(cfg *config.Config, cli client.IClient) (*Server, error) {
	s := &Server{
		operate: operate.NewService(cli),
		notice:  notice.NewService(cli),
		cfg:     cfg,
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
	operate_service.RegisterOperateServiceServer(s.serv, s.operate)
	operate_service.RegisterNoticeServiceServer(s.serv, s.notice)

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
		orm.CronStop()
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
