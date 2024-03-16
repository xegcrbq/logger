package logger

import (
	"context"
	"golang.org/x/mod/modfile"
	"log"
	pb "logger/grpc"
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
			_, err := lg.client.SendLogs(context.Background(), &lg.logs)
			if err != nil {
				log.Printf("Error send logs: %v", err)
				continue
			}
			lg.logs = pb.LogBatch{}
		}
	}
}
func forwardSendLogs() {
	_, err := lg.client.SendLogs(context.Background(), &lg.logs)
	if err != nil {
		log.Printf("Error send logs: %v", err)
	}
}
