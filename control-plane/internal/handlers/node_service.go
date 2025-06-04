package handlers

import (
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/db"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
)

type NodeService struct {
	pb.UnimplementedNodeServiceServer

	repo  repository.NodeRepository
	Store *db.DBStore
}

func NewNodeService(repo repository.NodeRepository) *NodeService {
	return &NodeService{repo: repo}
}
