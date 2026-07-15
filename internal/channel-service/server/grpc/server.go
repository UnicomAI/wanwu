package grpc

import (
	"context"
	"net"
	"time"

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	channel_service "github.com/UnicomAI/wanwu/api/proto/channel-service"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter"
	"github.com/UnicomAI/wanwu/internal/channel-service/chat"
	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	"github.com/UnicomAI/wanwu/internal/channel-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type Server struct {
	cfg             *config.Config
	serv            *grpc.Server
	cli             client.IClient
	channel         *ChannelService
	manager         *adapter.Manager
	handler         *chat.Handler
	qrCleanupStop   func() // 过期扫码会话清理协程的停止函数
	healthCheckStop func() // 通道心跳巡检协程的停止函数
}

func NewServer(cfg *config.Config, cli client.IClient) (*Server, error) {
	s := &Server{
		cfg: cfg,
		cli: cli,
	}

	// 创建适配器管理器
	mgr := adapter.NewManager(*cfg, cli)

	// 创建消息处理器
	chatHandler := chat.NewHandler(*cfg, cli, mgr)

	// 设置全局消息处理函数：平台消息 → chat handler → wanwu OpenAPI → 回复
	mgr.SetMessageHandler(chatHandler.HandlePlatformMessage)

	s.manager = mgr
	s.handler = chatHandler
	s.channel = NewChannelService(cfg, cli, mgr)

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	if s.serv != nil {
		return nil
	}

	s.serv = trace_util.NewGrpcTracerServer(
		[]grpc.UnaryServerInterceptor{trace_util.LoggingUnaryGRPC()},
		[]grpc.StreamServerInterceptor{trace_util.LoggingStreamGRPC()},
	)

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(s.serv, healthcheck)

	// register service
	channel_service.RegisterChannelServiceServer(s.serv, s.channel)

	// 启动适配器管理器（自动连接所有 enabled + loggedIn 的通道）
	go func() {
		if err := s.manager.StartAll(ctx); err != nil {
			log.Errorf("failed to start adapter manager: %v", err)
		}
	}()

	// 启动过期扫码会话定时清理（每 1 小时一次，随 ctx 退出而停止）
	s.qrCleanupStop = s.channel.qrMgr.StartCleanup(ctx, time.Hour)

	// 启动通道心跳巡检（每 2 分钟探测所有已启动通道，状态异常时回写 DB 供前端感知）
	s.healthCheckStop = s.manager.StartHealthCheck(ctx, 2*time.Minute)

	// listen
	lis, err := net.Listen("tcp", s.cfg.Server.GrpcEndpoint)
	if err != nil {
		return err
	}

	go func() {
		err = s.serv.Serve(lis)
		if err != nil {
			log.Fatalf("grpc server.Serve() failed, err: %v", err)
		}
	}()

	log.Infof("channel-service start grpc server at: %s", s.cfg.Server.GrpcEndpoint)
	return nil
}

func (s *Server) Stop(ctx context.Context) {
	if s.serv == nil {
		return
	}

	// 停止过期扫码会话清理协程
	if s.qrCleanupStop != nil {
		s.qrCleanupStop()
	}

	// 停止通道心跳巡检协程
	if s.healthCheckStop != nil {
		s.healthCheckStop()
	}

	// 停止所有适配器
	s.manager.StopAll()

	log.Infof("closing channel-service grpc server...")
	stopped := make(chan struct{})
	go func() {
		s.serv.GracefulStop()
		log.Infof("close channel-service grpc server gracefully")
		close(stopped)
	}()

	cancelCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	select {
	case <-cancelCtx.Done():
		s.serv.Stop()
		log.Errorf("close channel-service grpc server forced")
	case <-stopped:
	}
}
