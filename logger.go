package logger

import (
	"context"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"log"
	pb "logger/grpc"
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
	if ctxLogger := loggerFromContext(ctx); ctxLogger != nil {
		return ctxLogger
	}
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		ctx, span = otel.Tracer("lower").Start(ctx, "lower")
	}
	return &CLg{
		span:    &span,
		traceID: span.SpanContext().TraceID().String(),
		skip:    1,
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
