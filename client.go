package logger

import (
	pb "github.com/xegcrbq/logger/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(addr string) pb.LogServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		lg.lgCtx.Fatalf("Error create gRPC client with addr = %s: %v", addr, err)
	}
	return pb.NewLogServiceClient(conn)
}
