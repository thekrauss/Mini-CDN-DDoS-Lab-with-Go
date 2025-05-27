package handlers

import "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"

type NodeService struct {
	repo repository.NodeRepository
}

func NewNodeService(repo repository.NodeRepository) *NodeService {
	return &NodeService{repo: repo}
}
