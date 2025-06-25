package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/workflows"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *NodeService) BlacklistNode(ctx context.Context, req *pb.NodeID) (*emptypb.Empty, error) {
	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id requis")
	}

	// claims, err := auth.ExtractJWTFromContext(ctx)
	// if err != nil {
	// 	return nil, status.Errorf(codes.Unauthenticated, "Token invalide")
	// }

	// if err := s.CheckAdminPermissions(ctx, claims, PermManageNode); err != nil {
	// 	return nil, err
	// }

	node, err := s.Repo.GetNodeByID(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Node non trouvé: %v", err)
	}

	if node.IsBlacklisted {
		return &emptypb.Empty{}, nil // déjà blacklisté
	}

	err = s.Repo.SetNodeBlacklistStatus(ctx, req.NodeId, true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Impossible de blacklister le node: %v", err)
	}

	_ = pkg.RedisClient.Set(ctx, fmt.Sprintf("node:blacklist:%s", req.NodeId), "1", 0)

	log.Printf("[BLACKLIST] Node %s (%s) blacklisté avec succès", node.ID, node.IP)

	if s.TemporalClient != nil {
		workflowOptions := client.StartWorkflowOptions{
			ID:        fmt.Sprintf("audit-blacklist-%s-%d", req.NodeId, time.Now().Unix()),
			TaskQueue: "AUDIT_TASK_QUEUE",
		}

		Ip, userAgent := GetRequestMetadata(ctx)

		auditInput := workflows.AuditInput{
			Action:    "blacklist",
			TargetID:  node.ID,
			UserID:    claims.UserID.String(),
			TenantID:  node.TenantID,
			Role:      claims.Role,
			IPAddress: Ip,
			UserAgent: userAgent,
			Timestamp: time.Now(),
		}

		_, err := s.TemporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.AuditWorkflow, auditInput)
		if err != nil {
			log.Printf("Erreur lancement workflow d’audit : %v", err)
		}

	}
	return &emptypb.Empty{}, nil
}

func (s *NodeService) UnblacklistNode(ctx context.Context, req *pb.NodeID) (*emptypb.Empty, error) {
	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id requis")
	}

	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide")
	}

	// if err := s.CheckAdminPermissions(ctx, claims, PermManageNode); err != nil {
	// 	return nil, err
	// }

	node, err := s.Repo.GetNodeByID(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Node non trouvé: %v", err)
	}

	if !node.IsBlacklisted {
		return nil, status.Errorf(codes.FailedPrecondition, "Le nœud n’est pas blacklisté")
	}

	if err := s.Repo.SetNodeBlacklistStatus(ctx, req.NodeId, false); err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur suppression de blacklist: %v", err)
	}

	_ = pkg.RedisClient.Del(ctx, fmt.Sprintf("node:blacklist:%s", req.NodeId))

	log.Printf("[UNBLACKLIST] Node %s retiré de la blacklist", req.NodeId)

	if s.TemporalClient != nil {
		workflowOptions := client.StartWorkflowOptions{
			ID:        fmt.Sprintf("audit-unblacklist-%s-%d", req.NodeId, time.Now().Unix()),
			TaskQueue: "AUDIT_TASK_QUEUE",
		}

		ip, userAgent := GetRequestMetadata(ctx)

		auditInput := workflows.AuditInput{
			Action:    "unblacklist",
			TargetID:  node.ID,
			UserID:    claims.UserID.String(),
			TenantID:  node.TenantID,
			Role:      claims.Role,
			IPAddress: ip,
			UserAgent: userAgent,
			Timestamp: time.Now(),
		}

		_, err := s.TemporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.AuditWorkflow, auditInput)
		if err != nil {
			log.Printf("⚠ Erreur lancement workflow d’audit unblacklist : %v", err)
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *NodeService) ListBlacklistedNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	if req.TenantId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id requis")
	}

	// claims, err := auth.ExtractJWTFromContext(ctx)
	// if err != nil {
	// 	return nil, status.Errorf(codes.Unauthenticated, "Token invalide")
	// }

	// if err := s.CheckAdminPermissions(ctx, claims, PermManageNode); err != nil {
	// 	return nil, err
	// }

	nodes, err := s.Repo.ListBlacklistedNodes(ctx, req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur récupération nodes blacklistés: %v", err)
	}

	var pbNodes []*pb.Node
	for _, node := range nodes {
		pbNodes = append(pbNodes, &pb.Node{
			Id:              node.ID,
			Name:            node.Name,
			Ip:              node.IP,
			TenantId:        node.TenantID,
			Status:          node.Status,
			LastSeen:        node.LastSeen.Format(time.RFC3339),
			CreatedAt:       node.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       node.UpdatedAt.Format(time.RFC3339),
			Location:        node.Location,
			Provider:        node.Provider,
			SoftwareVersion: node.SoftwareVersion,
			IsBlacklisted:   node.IsBlacklisted,
			Tags:            node.Tags,
			Os:              node.OS,
		})
	}

	return &pb.ListNodesResponse{Nodes: pbNodes}, nil
}
