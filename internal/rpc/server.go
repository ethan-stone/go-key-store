package rpc

import (
	"context"
	"log"

	"github.com/ethan-stone/go-key-store/internal/service"
	"google.golang.org/grpc"
)

type RpcServer struct {
	UnimplementedStoreServiceServer
	storeService service.StoreService
}

func (s *RpcServer) Ping(_ context.Context, req *PingRequest) (*PingResponse, error) {
	log.Println("Ping request received.")
	return &PingResponse{Ok: true}, nil
}

func (s *RpcServer) Get(_ context.Context, req *GetRequest) (*GetResponse, error) {
	log.Printf("Get request received for key %s", req.GetKey())

	val, err := s.storeService.Get(req.GetKey())

	if err != nil {
		return nil, err
	}

	return &GetResponse{
		Key: req.GetKey(),
		Val: val,
		Ok:  true,
	}, nil
}

func (s *RpcServer) Put(_ context.Context, req *PutRequest) (*PutResponse, error) {
	log.Printf("Put request received for key %s", req.GetKey())

	err := s.storeService.Put(req.GetKey(), req.GetVal())

	if err != nil {
		return nil, err
	}

	return &PutResponse{
		Ok: true,
	}, nil
}

func (s *RpcServer) Delete(_ context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	log.Printf("Delete request received for key %s", req.GetKey())

	err := s.storeService.Delete(req.GetKey())

	if err != nil {
		return nil, err
	}

	return &DeleteResponse{
		Ok: true,
	}, nil
}

func NewRpcServer(storeService service.StoreService) *grpc.Server {
	grpcServer := grpc.NewServer()

	RegisterStoreServiceServer(grpcServer, &RpcServer{
		storeService: storeService,
	})

	return grpcServer
}
