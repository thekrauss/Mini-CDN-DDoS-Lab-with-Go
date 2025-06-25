package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/logger"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RegisterNode permet d'enregistrer un nouveau worker-node dans l'infrastructure.
//
// Cette méthode est appelée depuis l’interface d’administration ou via une API lors de l’installation initiale de l’agent.
//
// Étapes effectuées :
//  Validation des champs obligatoires (IP, hostname, tenant).
//  Vérification du format de l’adresse IP.
//  Authentification de l’administrateur via JWT extrait du contexte.
//  Validation des permissions .
//  Vérification de l’unicité de l’IP pour éviter les doublons.
//  Enregistrement du nœud dans PostgreSQL avec un ID UUID.
//  Mise en cache facultative du nœud (Redis).
//  Insertion d’un log d’audit pour traçabilité (action, user, IP, timestamp).
//  Contrôle du quota : max 20 nœuds actifs par tenant.
//
// Si l’enregistrement réussit, retourne le NodeID attribué.

func (s *NodeService) RegisterNode(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("RegisterNode: ID=%s, IP=%s", req.NodeId, req.Ip)

	if req.Ip == "" || req.Hostname == "" || req.IdTenant == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Les champs IP, hostname et tenant_id sont obligatoires")
	}

	parsedIP := net.ParseIP(req.Ip)
	if parsedIP == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Adresse IP invalide")
	}

	// authentification par contexte
	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}

	adminUser, err := pkg.GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		log.Printf("Impossible de récupérer les infos admin depuis Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "Impossible de récupérer les informations administrateur")
	}

	// if err := s.CheckAdminPermissions(ctx, claims, PermManageNode); err != nil {
	// 	return nil, err
	// }

	//  si l'IP est déjà utilisée
	exists, err := s.Repo.IsIPAlreadyRegistered(ctx, req.Ip)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}
	if exists {
		return nil, status.Errorf(codes.AlreadyExists, "L'adresse IP est déja utilisée")
	}

	//  nouveau noeud
	node := &repository.Node{
		ID:              uuid.New().String(),
		Name:            req.Hostname,
		IP:              req.Ip,
		TenantID:        req.IdTenant,
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

	maxNodes := s.Config.MonitoringEtat.MaxNodesPerTenant
	count, _ := s.Repo.CountActiveNodes(ctx, node.TenantID, 24*time.Hour)
	if count >= maxNodes {
		return nil, status.Errorf(codes.ResourceExhausted, "Quota de nodes atteint pour ce tenant")
	}

	log.Printf("[DEBUG] Max nodes per tenant: %d", maxNodes)

	go func() {
		if err := pkg.CacheNode(ctx, node); err != nil {
			logger.Log.Warn("Échec mise en cache node", zap.Error(err))
		}
	}()

	err = s.Repo.CreateNode(ctx, node)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "échec de l'enregistrement du nœud: %v", err)
	}

	ip, userAgent := GetRequestMetadata(ctx)

	audit := &repository.AuditLog{
		ID:        uuid.New(),
		UserID:    claims.UserID,
		Role:      adminUser.Role,
		Action:    "RegisterNode",
		Target:    node.ID,
		Details:   fmt.Sprintf("Node enregistré avec IP=%s, Hostname=%s", node.IP, node.Name),
		IPAddress: ip,
		UserAgent: userAgent,
		TenantID:  uuid.MustParse(adminUser.TenantID),
		Timestamp: time.Now(),
	}

	if err := s.Repo.InsertAuditLog(ctx, audit); err != nil {
		logger.Log.Warn("Échec insertion audit log", zap.Error(err))
	}

	return &pb.RegisterResponse{
		Message: "Nœud enregistré avec succès",
		NodeId:  node.ID,
	}, nil
}

func (s *NodeService) UpdateNodeMetadata(ctx context.Context, req *pb.UpdateNodeRequest) (*emptypb.Empty, error) {
	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id requis")
	}

	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "token invalide")
	}

	node, err := s.Repo.GetNodeByID(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "nœud non trouvé: %v", err)
	}

	adminUser, err := pkg.GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		log.Printf("Erreur récupération utilisateur Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur accès utilisateur")
	}

	// if err := s.CheckAdminPermissions(ctx, claims, PermManageNode); err != nil {
	// 	return nil, err
	// }

	if adminUser.TenantID != node.TenantID && adminUser.Role != "GLOBAL_ADMIN" {
		return nil, status.Errorf(codes.PermissionDenied, "modification interdite sur ce nœud")
	}

	err = s.Repo.UpdateNodeMetadata(ctx, req.NodeId, req.Name, req.Ip, req.Tags)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur mise à jour node: %v", err)
	}

	// audit Log
	tagList := make([]string, 0, len(req.Tags))
	for k, v := range req.Tags {
		tagList = append(tagList, fmt.Sprintf("%s=%s", k, v))
	}

	ip, userAgent := GetRequestMetadata(ctx)
	details := fmt.Sprintf("Mise à jour nœud : nom=%s, ip=%s, tags=[%s]", req.Name, req.Ip, strings.Join(tagList, ", "))

	logEntry := &repository.AuditLog{
		ID:        uuid.New(),
		UserID:    claims.UserID,
		Role:      claims.Role,
		Action:    "UpdateNodeMetadata",
		Target:    req.NodeId,
		Details:   details,
		Timestamp: time.Now(),
		TenantID:  uuid.MustParse(node.TenantID),
		IPAddress: ip,
		UserAgent: userAgent,
	}

	_ = s.Repo.InsertAuditLog(ctx, logEntry)

	return &emptypb.Empty{}, nil
}
