package rpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

type RpcServer struct {
	UnimplementedStoreServiceServer
}

func (s *RpcServer) Ping(_ context.Context, req *PingRequest) (*PingResponse, error) {
	log.Println("Ping request received.")
	return &PingResponse{Ok: true}, nil
}

func NewRpcServer() *grpc.Server {
	grpcServer := grpc.NewServer()

	RegisterStoreServiceServer(grpcServer, &RpcServer{})

	return grpcServer
}
