package handlers

import (
	"context"

	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
)

func (s *NodeService) RegisterNode(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{Message: "OK"}, nil
}
