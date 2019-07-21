package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/tormaroe/eightlegs-project/meshagent/api"
)

type meshAgentServer struct {
}

func (s *meshAgentServer) MyStatusUpdate(ctx context.Context, status *api.MeshServiceStatus) (*api.MeshServiceStatusResponse, error) {
	log.Printf("Received service status for %s", status.GetServiceUuid())
	log.Printf("Service type: %s", status.GetServiceType())
	return &api.MeshServiceStatusResponse{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 50710))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	api.RegisterMeshAgentServer(grpcServer, &meshAgentServer{})
	grpcServer.Serve(lis)

}
