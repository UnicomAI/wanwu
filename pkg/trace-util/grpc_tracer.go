package trace_util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	headlessServiceSchema = "dns:///"
	maxMsgSize            = 1024 * 1024 * 1024 * 4 // 4G
)

func NewGrpcTracerServer(interceptors []grpc.UnaryServerInterceptor, streamInterceptors []grpc.StreamServerInterceptor) *grpc.Server {
	// init
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Errorf("[PANIC] %v\n%v", p, string(debug.Stack()))
			return status.Error(codes.Internal, fmt.Sprintf("panic: %v", p))
		}),
	}
	var interceptorList []grpc.UnaryServerInterceptor
	interceptorList = append(interceptorList, grpc_recovery.UnaryServerInterceptor(opts...))
	if len(interceptors) > 0 {
		interceptorList = append(interceptorList, interceptors...)
	}
	var streamInterceptorList []grpc.StreamServerInterceptor
	streamInterceptorList = append(streamInterceptorList, grpc_recovery.StreamServerInterceptor(opts...))
	if len(streamInterceptors) > 0 {
		streamInterceptorList = append(streamInterceptorList, streamInterceptors...)
	}
	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(interceptorList...),
		grpc.ChainStreamInterceptor(streamInterceptorList...),
	}
	//if len(serverOptionList) > 0 {
	//	serverOptions = append(serverOptions, serverOptionList...)
	//}
	return grpc.NewServer(serverOptions...)
}

func NewGrpcTracerConn(host string, intercepts []grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
			grpc.MaxCallSendMsgSize(maxMsgSize)),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	if len(intercepts) > 0 {
		options = append(options, grpc.WithChainUnaryInterceptor(intercepts...))
	}
	conn, err := grpc.NewClient(headlessServiceSchema+host,
		options...,
	)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func LoggingUnaryGRPC() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		startTime := time.Now()
		requestId := GetTraceID(ctx)

		// 记录请求
		reqBuf := new(bytes.Buffer)
		if err := json.NewEncoder(reqBuf).Encode(req); err != nil {
			log.Errorf("[Request ID: %s] Request Method %s | Failed to encode request: %v", requestId, info.FullMethod, err)
		}
		log.Infof("[Request ID: %s] Request Method %s | Request Body: %s", requestId, info.FullMethod, reqBuf.String())
		// 将请求ID添加到上下文中，以便下游服务也可以访问它
		//ctx = context.WithValue(ctx, "request_id", requestId)

		// 调用下一个handler
		resp, err := handler(ctx, req)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("[Request ID: %s] Error handling request: %v", requestId, err)
			return nil, err
		}

		// 记录响应
		respBuf := new(bytes.Buffer)
		if err := json.NewEncoder(respBuf).Encode(resp); err != nil {
			log.Errorf("[Request ID: %s] Failed to encode response: %v", requestId, err)
		}
		log.Infof("[Request ID: %s] Request Method %s | Request Duration: %s, Response Body: %s", requestId, info.FullMethod, duration, respBuf.String())

		return resp, err
	}
}

func LoggingStreamGRPC() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		// 从上下文中获取 SpanContext（如果存在的话）
		ctx := ss.Context()
		requestId := GetTraceID(ctx)

		log.Infof("[Stream Request ID: %s] Request Method %s | Start", requestId, info.FullMethod)

		// 调用下一个handler
		err := handler(srv, ss)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("[Stream Request ID: %s] Request Method %s | Error: %v, Duration: %s", requestId, info.FullMethod, err, duration)
			return err
		}

		log.Infof("[Stream Request ID: %s] Request Method %s | Duration: %s", requestId, info.FullMethod, duration)
		return nil
	}
}
