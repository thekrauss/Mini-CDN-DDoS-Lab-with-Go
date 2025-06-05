package handlers

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *NodeService) RegisterNode(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("RegisterNode: ID=%s, IP=%s", req.NodeId, req.Ip)

	// authentification par contexte
	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}

	tenantID := claims.ID
	if tenantID == "" {
		return nil, status.Errorf(codes.PermissionDenied, "no tenant assigned")
	}

	//  si l'IP est déjà utilisée
	exists, err := s.repo.IsIPAlreadyRegistered(ctx, req.Ip)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}
	if exists {
		return nil, status.Errorf(codes.AlreadyExists, "this IP is already in use")
	}

	// Créer un nouveau noeud
	node := &repository.Node{
		ID:              uuid.New().String(),
		Name:            req.Hostname,
		IP:              req.Ip,
		TenantID:        tenantID,
		Status:          string(repository.NodeOnline),
		LastSeen:        time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Location:        req.Location,
		Provider:        req.Provider,
		SoftwareVersion: req.Version,
		IsBlacklisted:   false,
		Tags:            req.Tags,
	}

	err = s.repo.CreateNode(ctx, node)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register node: %v", err)
	}

	return &pb.RegisterResponse{
		Message: "Node registered successfully",
		NodeId:  node.ID,
	}, nil
}

func (s *NodeService) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	log.Printf("→ Ping from node %s: CPU=%.2f MEM=%.2f", req.NodeId, req.Cpu, req.Memory)
	return &pb.PingResponse{Status: "ok"}, nil
}

func (s *NodeService) SendMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsResponse, error) {
	log.Printf("→ Metrics from node %s: CPU=%.2f MEM=%.2f", req.NodeId, req.Cpu, req.Memory)
	return &pb.MetricsResponse{Status: "received"}, nil
}
