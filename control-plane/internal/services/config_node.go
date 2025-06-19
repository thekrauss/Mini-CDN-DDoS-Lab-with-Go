package services

import (
	"context"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *NodeService) GetNodeConfig(ctx context.Context, req *pb.GetNodeConfigRequest) (*pb.GetNodeConfigResponse, error) {
	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id requis")
	}

	_, err := s.Repo.GetNodeByID(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Node introuvable")
	}

	// Chargement depuis DB ou valeurs par défaut de la config globale
	cfg := config.AppConfig.Agent
	nodeCfg, err := s.Repo.GetNodeConfig(ctx, req.NodeId)
	if err != nil {
		// fallback par défaut
		nodeCfg = &repository.NodeConfig{
			NodeID:          req.NodeId,
			PingInterval:    cfg.DefaultPingInterval,
			MetricsInterval: cfg.DefaultMetricsInterval,
			DynamicConfig:   cfg.EnableDynamicConfig,
			CustomLabels:    map[string]string{},
		}

	}

	return &pb.GetNodeConfigResponse{
		NodeId:               nodeCfg.NodeID,
		PingInterval:         int32(nodeCfg.PingInterval),
		MetricsInterval:      int32(nodeCfg.MetricsInterval),
		DynamicConfigEnabled: nodeCfg.DynamicConfig,
		CustomLabels:         nodeCfg.CustomLabels,
	}, nil
}
