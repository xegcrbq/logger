package logger

import (
	"context"
	pb "github.com/fle4a/logger/grpc"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"os"
	"sync"
	"time"
)

type logger struct {
	client      pb.LogServiceClient
	logs        pb.LogBatch
	serviceName string
	mu          sync.Mutex
	lgCtx       ICLg
}

type CtxOptions struct {
	Tracer   trace.Tracer
	SpanName string
}

var lg *logger
var zl zerolog.Logger

func Connect(addr string) {
	moscowLocation, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		lg.lgCtx.Info("Failed to load Moscow timezone")
	}
	var consoleWriterPtr *zerolog.ConsoleWriter
	zerolog.TimeFieldFormat = time.StampMilli
	zerolog.ErrorFieldName = zerolog.MessageFieldName
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(moscowLocation)
	}

	consoleWriterPtr = &zerolog.ConsoleWriter{Out: os.Stderr,
		TimeFormat:    time.TimeOnly,
		FieldsExclude: []string{"stack"},
		NoColor:       false,
	}
	zl = zerolog.New(consoleWriterPtr).
		With().
		Timestamp().
		Caller().
		Stack().
		Logger()

	lg = &logger{
		client:      NewClient(addr),
		logs:        pb.LogBatch{},
		serviceName: getServiceName(),
		lgCtx:       Ctx(context.Background()),
	}
	go sendLogs()
}

func Ctx(ctx context.Context, opts ...string) ICLg {
	if len(opts) == 2 {
		ctx, span := otel.Tracer(opts[1]).Start(ctx, opts[0])
		return &CtxLogger{
			span:    &span,
			traceID: span.SpanContext().TraceID().String(),
			skip:    1,
			Context: ctx,
			extra:   make(map[string]string),
		}
	}
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		ctx, span = otel.Tracer("logger").Start(ctx, "logger")
	}
	return &CtxLogger{
		span:    &span,
		traceID: span.SpanContext().TraceID().String(),
		skip:    1,
		Context: ctx,
		extra:   make(map[string]string),
	}
}
