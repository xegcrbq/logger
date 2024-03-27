package logger

import (
	"context"
	pb "github.com/fle4a/logger/grpc"
	"github.com/rs/zerolog"
	"golang.org/x/mod/modfile"
	"google.golang.org/grpc/metadata"
	"os"
	"time"
)

func getServiceName() string {
	data, _ := os.ReadFile("go.mod")
	return modfile.ModulePath(data)
}

func sendLogs() {
	for {
		select {
		case <-time.After(time.Second * time.Duration(SENDPERIOD)):
			if len(lg.logs.Logs) > 0 && lg.client != nil {
				md := metadata.Pairs("serviceName", getServiceName())
				ctx := metadata.NewOutgoingContext(context.Background(), md)
				_, err := lg.client.SendLogs(ctx, &lg.logs)
				if err != nil {
					lg.lgCtx.Errorf("Error send logs: %v", err)
					continue
				}
				lg.logs = pb.LogBatch{}
			} else if len(lg.logs.Logs) > 1000 {
				lg.logs = pb.LogBatch{}
			}
		}
	}
}

func forwardSendLogs() {
	if len(lg.logs.Logs) > 0 && lg.client != nil {
		md := metadata.Pairs("serviceName", lg.serviceName)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		_, err := lg.client.SendLogs(ctx, &lg.logs)
		if err != nil {
			lg.lgCtx.Errorf("Error send logs: %v", err)
		}
	}
}

func getZeroLogger() zerolog.Logger {
	moscowLocation, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		moscowLocation = time.UTC
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
	return zl
}
