package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *NodeService) ListNodesByTenant(ctx context.Context, req *pb.TenantRequest) (*pb.NodeListResponse, error) {
	if req.TenantId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "tenant_id requis")

	}

	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}

	adminUser, err := pkg.GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		log.Printf("Impossible de récupérer les infos admin depuis Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "Impossible de récupérer les informations administrateur")
	}

	if err := s.CheckAdminPermissions(ctx, claims, PermReadNode); err != nil {
		return nil, err
	}

	if adminUser.TenantID != req.TenantId {
		return nil, status.Errorf(codes.PermissionDenied, "accès interdit à ce tenant")
	}

	nodes, err := s.Repo.ListNodesByTenant(ctx, req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur récupération des nœuds: %v", err)
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
			Os:              node.OS,
			SoftwareVersion: node.SoftwareVersion,
			Provider:        node.Provider,
			IsBlacklisted:   node.IsBlacklisted,
			Tags:            node.Tags,
		})
	}

	return &pb.NodeListResponse{
		Nodes: pbNodes,
	}, nil

}

func (s *NodeService) SetNodeStatus(ctx context.Context, req *pb.NodeStatusRequest) (*emptypb.Empty, error) {
	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id requis")
	}

	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide")
	}

	node, err := s.Repo.GetNodeByID(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Nœud introuvable: %v", err)
	}

	adminUser, err := pkg.GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur autorisation")
	}
	if adminUser.TenantID != node.TenantID && adminUser.Role != "GLOBAL_ADMIN" {
		return nil, status.Errorf(codes.PermissionDenied, "Accès interdit à ce nœud")
	}

	//enum → string
	statusStr, err := NodeStatusToString(req.Status)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	//met à jour PostgreSQL
	if err := s.Repo.SetNodeStatus(ctx, req.NodeId, statusStr); err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur mise à jour du statut: %v", err)
	}

	// redis
	keyStatus := fmt.Sprintf("node:status:%s", req.NodeId)
	if statusStr == string(repository.NodeOffline) {
		_ = pkg.RedisClient.Del(ctx, keyStatus)
	} else {
		_ = pkg.RedisClient.Set(ctx, keyStatus, statusStr, 2*time.Minute).Err()
	}

	ip, userAgent := GetRequestMetadata(ctx)

	// journalisation audit
	logEntry := &repository.AuditLog{
		ID:        uuid.New(),
		UserID:    claims.UserID,
		Role:      claims.Role,
		Action:    "SetNodeStatus",
		Target:    req.NodeId,
		Details:   fmt.Sprintf("Changement du statut en %s", statusStr),
		Timestamp: time.Now(),
		TenantID:  uuid.MustParse(node.TenantID),
		IPAddress: ip,
		UserAgent: userAgent,
	}
	_ = s.Repo.InsertAuditLog(ctx, logEntry)

	return &emptypb.Empty{}, nil
}

func NodeStatusToString(status pb.NodeStatus) (string, error) {
	switch status {
	case pb.NodeStatus_NODE_ONLINE:
		return "online", nil
	case pb.NodeStatus_NODE_OFFLINE:
		return "offline", nil
	case pb.NodeStatus_NODE_DEGRADED:
		return "degraded", nil
	default:
		return "", fmt.Errorf("statut inconnu: %v", status)
	}
}
