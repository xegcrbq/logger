package logger

import (
	pb "github.com/fle4a/logger/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func NewClient(addr string) pb.LogServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error create gRPC client with addr = %s: %v", addr, err)
	}
	return pb.NewLogServiceClient(conn)
}
