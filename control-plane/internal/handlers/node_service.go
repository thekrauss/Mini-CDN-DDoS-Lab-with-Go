package handlers

import (
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/db"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"

	authpb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
)

type NodeService struct {
	pb.UnimplementedNodeServiceServer
	Repo       repository.NodeRepository
	Store      *db.DBStore
	AuthClient authpb.AuthServiceClient
}

func NewNodeService(repo repository.NodeRepository) *NodeService {
	return &NodeService{Repo: repo}
}
