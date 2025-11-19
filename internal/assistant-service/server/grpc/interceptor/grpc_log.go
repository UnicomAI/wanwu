package interceptor

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func LoggingUnaryGRPC() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		startTime := time.Now()
		requestId := uuid.New().String()

		// Logging request
		reqBuf := new(bytes.Buffer)
		if err := json.NewEncoder(reqBuf).Encode(req); err != nil {
			log.Errorf("[Request ID: %s] Request Method %s | Failed to encode request: %v", requestId, info.FullMethod, err)
		}
		log.Infof("[Request ID: %s] Request Method %s | Request Body: %s", requestId, info.FullMethod, reqBuf.String())
		// Add the request ID to the context so that downstream services can also access it
		//ctx = context.WithValue(ctx, "request_id", requestId)

		// Call next handler
		resp, err := handler(ctx, req)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		if err != nil {
			log.Errorf("[Request ID: %s] Error handling request: %v", requestId, err)
			return nil, err
		}

		// Log response
		respBuf := new(bytes.Buffer)
		if err := json.NewEncoder(respBuf).Encode(resp); err != nil {
			log.Errorf("[Request ID: %s] Failed to encode response: %v", requestId, err)
		}
		log.Infof("[Request ID: %s] Request Method %s | Request Duration: %s, Response Body: %s", requestId, info.FullMethod, duration, respBuf.String())

		return resp, err
	}
}
