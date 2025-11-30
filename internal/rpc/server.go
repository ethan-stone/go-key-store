package rpc

import (
	"context"
	"fmt"
	"log"

	"github.com/ethan-stone/go-key-store/internal/configuration"
	"github.com/ethan-stone/go-key-store/internal/local_store"
	"github.com/ethan-stone/go-key-store/internal/wal"
	"google.golang.org/grpc"
)

type RpcServer struct {
	UnimplementedStoreServiceServer
	storeService     *local_store.LocalKeyValueStore
	rpcClientManager RpcClientManager
	configManager    configuration.ConfigurationManager
}

func (s *RpcServer) Ping(_ context.Context, req *PingRequest) (*PingResponse, error) {
	log.Println("Ping request received.")
	return &PingResponse{Ok: true}, nil
}

func (s *RpcServer) Get(_ context.Context, req *GetRequest) (*GetResponse, error) {
	log.Printf("Get request received for key %s", req.GetKey())

	result, err := s.storeService.Get(req.GetKey())

	if err != nil {
		return nil, err
	}

	if !result.Ok {
		return &GetResponse{
			Key: req.GetKey(),
			Val: "",
			Ok:  false,
		}, nil
	}

	return &GetResponse{
		Key: req.GetKey(),
		Val: result.Val,
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

func (s *RpcServer) AppendWalEntry(_ context.Context, req *AppendWalEntryRequest) (*AppendWalEntryResponse, error) {
	log.Printf("Received AppendWalEntry request for sequence number %d", req.GetWalEntry().GetSequenceNumber())

	keyBytes := make([]byte, len(req.GetWalEntry().GetKeyBytes()))
	copy(keyBytes, req.GetWalEntry().GetKeyBytes())

	valueBytes := make([]byte, len(req.GetWalEntry().GetValueBytes()))
	copy(valueBytes, req.GetWalEntry().GetValueBytes())

	walEntry := &wal.WalEntryWrite{
		OpType:      byte(req.GetWalEntry().GetOpType()),
		KeyLength:   req.GetWalEntry().GetKeyLength(),
		ValueLength: req.GetWalEntry().GetValueLength(),
		KeyBytes:    &keyBytes,
		ValueBytes:  &valueBytes,
	}

	sequenceNumber, err := s.storeService.GetWalWriter().Write(walEntry)

	if sequenceNumber != req.GetWalEntry().GetSequenceNumber() {
		return nil, fmt.Errorf("sequence number mismatch. expected %d, got %d", req.GetWalEntry().GetSequenceNumber(), sequenceNumber)
	}

	if err != nil {
		return nil, err
	}

	return &AppendWalEntryResponse{
		Ok: true,
	}, nil
}

func NewRpcServer(storeService *local_store.LocalKeyValueStore, configManager configuration.ConfigurationManager, rpcClientManager RpcClientManager) *grpc.Server {
	grpcServer := grpc.NewServer()

	RegisterStoreServiceServer(grpcServer, &RpcServer{
		storeService:     storeService,
		rpcClientManager: rpcClientManager,
		configManager:    configManager,
	})

	return grpcServer
}
