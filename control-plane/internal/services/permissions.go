package services

import (
	"context"
	"fmt"
	"log"
	"time"

	authpb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/db"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/ws"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeService struct {
	pb.UnimplementedNodeServiceServer
	Repo       repository.NodeRepository
	Store      *db.DBStore
	AuthClient authpb.AuthServiceClient
	Config     config.Config
	Hub        *ws.Hub
}

const (
	PermManageNode        = "MANAGE_NODE"
	PermReadNode          = "READ_NODE"
	PermPingNode          = "PING_NODE"
	PermSendMetrics       = "SEND_METRICS"
	PermReadMetrics       = "READ_METRICS"
	PermReadAuditLogs     = "READ_AUDIT_LOGS"
	PermManageConfig      = "MANAGE_CONFIG"
	PermManageTenant      = "MANAGE_TENANT"
	PermReadTenant        = "READ_TENANT"
	PermManageUsers       = "MANAGE_USERS"
	PermManagePermissions = "MANAGE_PERMISSIONS"
	PermRestartService    = "RESTART_NODE_SERVICE"
	PermStopService       = "STOP_NODE_SERVICE"
	PermUpdateConfig      = "UPDATE_NODE_CONFIG"
)

func (s *NodeService) CheckAdminPermissions(ctx context.Context, claims *auth.Claims, permission string) error {

	// cas classiques
	if config.IsSuperAdmin(claims.Role) {
		return nil
	}

	if config.IsTenantAdmin(claims.Role) {
		hasPerm, err := s.Permission(ctx, claims.UserID.String(), permission)
		if err != nil {
			return status.Errorf(codes.Internal, "Erreur interne lors de la vérification des permissions")
		}
		if !hasPerm {
			return status.Errorf(codes.PermissionDenied, "Vous n'avez pas la permission d'effectuer cette action")
		}
		return nil
	}

	return status.Errorf(codes.PermissionDenied, "Accès refusé")
}

func (s *NodeService) Permission(ctx context.Context, userID, requiredPermission string) (bool, error) {
	permissionsKey := fmt.Sprintf("cdn-permissions:%s", userID)

	//  dans Redis
	exists, err := pkg.RedisClient.SIsMember(ctx, permissionsKey, requiredPermission).Result()
	if err == nil && exists {
		log.Printf("Permission '%s' trouvée dans Redis pour %s", requiredPermission, userID)
		return true, nil
	}

	//  gRPC vers auth-service
	res, err := s.AuthClient.HasPermission(ctx, &authpb.HasPermissionRequest{
		UserId:     userID,
		Permission: requiredPermission,
	})
	if err != nil {
		log.Printf("Erreur gRPC HasPermission: %v", err)
		return false, err
	}

	if res.Allowed {
		// met en cache pour la prochaine fois
		if err := CachCdnPermissionsInRedis(ctx, userID, []string{requiredPermission}); err != nil {
			log.Printf("Impossible de mettre en cache Redis : %v", err)
		}
		return true, nil
	}

	log.Printf("Permission '%s' refusée pour %s", requiredPermission, userID)
	return false, nil
}

func CachCdnPermissionsInRedis(ctx context.Context, userID string, permissions []string) error {
	permissionsKey := fmt.Sprintf("cdn-permissions:%s", userID)
	err := pkg.RedisClient.SAdd(ctx, permissionsKey, permissions).Err()
	if err != nil {
		return err
	}
	pkg.RedisClient.Expire(ctx, permissionsKey, 24*time.Hour)
	return nil
}

func RemoveCachedCdnPermissions(ctx context.Context, userID string) error {
	permissionsKey := fmt.Sprintf("cdn-permissions:%s", userID)
	err := pkg.RedisClient.Del(ctx, permissionsKey).Err()
	if err != nil {
		return err
	}
	log.Printf("Cache Redis supprimé pour %s", permissionsKey)
	return nil
}
