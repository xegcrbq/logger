package logger

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	pg "logger/grpc"
)

func NewClient(addr string) pg.LogServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error create gRPC client with addr = %s: %v", addr, err)
	}
	return pg.NewLogServiceClient(conn)
}
