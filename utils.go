package logger

import (
	"context"
	pb "github.com/fle4a/logger/grpc"
	"golang.org/x/mod/modfile"
	"google.golang.org/grpc/metadata"
	"log"
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
			if len(lg.logs.Logs) > 0 {
				md := metadata.Pairs("serviceName", getServiceName())
				ctx := metadata.NewOutgoingContext(context.Background(), md)
				_, err := lg.client.SendLogs(ctx, &lg.logs)
				if err != nil {
					log.Printf("Error send logs: %v", err)
					continue
				}
				lg.logs = pb.LogBatch{}
			}
		}
	}
}
func forwardSendLogs() {
	md := metadata.Pairs("serviceName", lg.serviceName)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := lg.client.SendLogs(ctx, &lg.logs)
	if err != nil {
		log.Printf("Error send logs: %v", err)
	}
}
