package tracer

import (
	"context"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
)

var wanWuTracer = WanWuTracer{}

type WanWuTracer struct {
}

func init() {
	pkg.AddContainer(wanWuTracer)
}

func (c WanWuTracer) LoadType() string {
	return "tracer"
}

func (c WanWuTracer) Load() error {
	return trace_util.InitTracer("knowledge-service")
}

func (c WanWuTracer) StopPriority() int {
	return pkg.TracePriority
}

func (c WanWuTracer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	trace_util.ShutdownTracer(ctx)
	return nil
}
