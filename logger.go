package logger

import (
	"context"
	pb "github.com/fle4a/logger/grpc"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sync"
)

type logger struct {
	client      pb.LogServiceClient
	logs        pb.LogBatch
	serviceName string
	mu          sync.Mutex
}

type CtxOptions struct {
	Tracer   trace.Tracer
	SpanName string
}

var lg *logger
var zl zerolog.Logger

func Connect(addr string) {
	zl = zerolog.New(os.Stderr).With().Timestamp().Logger()
	lg = &logger{
		client:      NewClient(addr),
		logs:        pb.LogBatch{},
		serviceName: getServiceName(),
	}
	go sendLogs()
}

func Ctx(ctx context.Context) ICLg {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		ctx, span = otel.Tracer("lower").Start(ctx, "lower")
	}
	return &CLg{
		span:    &span,
		traceID: span.SpanContext().TraceID().String(),
		skip:    1,
		Context: ctx,
	}
}

func CtxWithSpan(ctx context.Context, opts CtxOptions) ICLg {
	if opts.Tracer == nil {
		opts.Tracer = otel.Tracer("logger")
	}
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		ctx, span = opts.Tracer.Start(ctx, "lower")
	}
	ctx, span = opts.Tracer.Start(ctx, opts.SpanName)
	return &CLg{
		span:    &span,
		traceID: span.SpanContext().TraceID().String(),
		skip:    1,
		Context: ctx,
	}

}

func Serve(addr string, server pb.LogServiceServer) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error listen on addr = %s: %v", addr, err)
	}
	gRPCServer := grpc.NewServer()
	pb.RegisterLogServiceServer(gRPCServer, server)
	go func() {
		if err = gRPCServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC on addr = %s: %v", addr, err)
		}
	}()
}
