package logger

import (
	"context"
	pb "github.com/fle4a/logger/grpc"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"sync"
)

type logger struct {
	client      pb.LogServiceClient
	logs        pb.LogBatch
	serviceName string
	mu          *sync.Mutex
	lgCtx       ICLg
}

type CtxOptions struct {
	Tracer   trace.Tracer
	SpanName string
}

var lg *logger
var zl zerolog.Logger

func init() {
	zl = getZeroLogger()
	go sendLogs()
	lg = &logger{
		logs:        pb.LogBatch{},
		serviceName: getServiceName(),
		mu:          &sync.Mutex{},
		lgCtx:       Ctx(context.Background()),
	}
}

func Connect(addr string) {
	lg.client = NewClient(addr)
}

func Ctx(ctx context.Context, opts ...string) ICLg {
	var span trace.Span
	if len(opts) == 2 {
		ctx, span = otel.Tracer(opts[1]).Start(ctx, opts[0])
	} else if len(opts) == 1 {
		ctx, span = otel.Tracer("logger").Start(ctx, opts[0])
	} else {
		span = trace.SpanFromContext(ctx)
		if !span.SpanContext().IsValid() {
			ctx, span = otel.Tracer("logger").Start(ctx, "logger")
		}
	}
	return &CtxLogger{
		span:    &span,
		traceID: span.SpanContext().TraceID().String(),
		skip:    1,
		Context: ctx,
		extra:   make(map[string]string),
	}
}
