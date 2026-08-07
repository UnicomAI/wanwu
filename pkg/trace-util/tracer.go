package trace_util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/UnicomAI/wanwu/api/proto/common"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	utils "github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/proto"

	go_redis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	TraceUserPrefix       = "trace_user:"
	TraceJaegerEnable     = "JAEGER_ENABLE"
	envJaegerOTLPEndpoint = "JAEGER_OTLP_ENDPOINT"
)

var _tracer = &Tracer{}

type Tracer struct {
	tp *sdktrace.TracerProvider
}

// InitTracer initializes the TracerProvider.
// When JAEGER_ENABLE is unset or false: no exporter, spans are discarded (backward compatible).
// When JAEGER_ENABLE=true + JAEGER_OTLP_ENDPOINT: exports spans via OTLP HTTP to Jaeger.
func InitTracer(serviceName string) error {
	enabled, _ := strconv.ParseBool(os.Getenv(TraceJaegerEnable))

	var tp *sdktrace.TracerProvider
	if enabled {
		endpoint := os.Getenv(envJaegerOTLPEndpoint)
		if endpoint == "" {
			return errors.New("JAEGER_ENABLE=1 but JAEGER_OTLP_ENDPOINT is empty")
		}
		var err error
		tp, err = initOTLPTracerProvider(serviceName, endpoint)
		if err != nil {
			return err
		}
		log.Infof("[trace] tracer initialized with OTLP exporter, endpoint=%s, service=%s", endpoint, serviceName)
	} else {
		tp = initDefaultTracerProvider()
		log.Infof("[trace] tracer initialized (no exporter, spans discarded)")
	}

	_tracer.tp = tp
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return nil
}

// ShutdownTracer flushes remaining spans before process exit.
func ShutdownTracer(ctx context.Context) {
	if _tracer.tp != nil {
		if err := _tracer.tp.Shutdown(ctx); err != nil {
			log.Errorf("[trace] tracer shutdown error: %v", err)
		}
	}
}

func GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	return spanCtx.TraceID().String()
}

// IsTraceContextValid reports whether ctx carries a valid OpenTelemetry trace.
func IsTraceContextValid(ctx context.Context) bool {
	return trace.SpanContextFromContext(ctx).IsValid()
}

func GetSpanID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	return spanCtx.SpanID().String()
}

// GetTraceUser retrieves trace user info from Redis by traceID.
func GetTraceUser(ctx context.Context) (traceInfo *common.TraceInfo, err error) {
	defer utils.PrintPanicStackWithCall(func(panicOccur bool, recoverError error) {
		if panicOccur {
			traceInfo = nil
			err = recoverError
		}
	})
	traceID := GetTraceID(ctx)
	traceKey := TraceUserKey(traceID)
	result := redis.OP().Cli().Get(ctx, traceKey)
	if result.Err() != nil {
		if errors.Is(result.Err(), go_redis.Nil) {
			return nil, nil
		}
		log.Errorf("get trace user failed: %v", result.Err())
		return nil, result.Err()
	}
	data, err := result.Result()
	if err != nil {
		log.Errorf("get trace user result failed: %v", err)
		return nil, err
	}
	traceInfo = &common.TraceInfo{}
	err = proto.Unmarshal([]byte(data), traceInfo)
	if err != nil {
		log.Errorf("unmarshal trace user failed: %v", err)
		return nil, err
	}
	return traceInfo, nil
}

// TraceUserKey returns the Redis key for trace user info.
func TraceUserKey(traceID string) string {
	return TraceUserPrefix + traceID
}

// initDefaultTracerProvider creates a TracerProvider without exporter (fallback mode).
func initDefaultTracerProvider() *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
}

// initOTLPTracerProvider creates a TracerProvider with OTLP HTTP exporter for Jaeger.
func initOTLPTracerProvider(serviceName, endpoint string) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP HTTP exporter failed: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		log.Warnf("[trace] resource.Merge failed: %v, using fallback resource", err)
		res = resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	), nil
}
