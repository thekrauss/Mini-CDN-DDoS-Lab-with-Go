package services

import (
	"context"
	"fmt"
	"log"
	"time"

	authpb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/db"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeService struct {
	Repo       repository.NodeRepository
	Store      *db.DBStore
	AuthClient authpb.AuthServiceClient
}

func (s *NodeService) CheckAdminPermissions(ctx context.Context, claims *auth.Claims, tenantID, permission string) error {

	// Cas classiques
	if config.IsRoleA(claims.Role) {
		return nil
	}

	if config.IsRoleB(claims.Role) {
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
	exists, err := RedisClient.SIsMember(ctx, permissionsKey, requiredPermission).Result()
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
		// Mise en cache pour la prochaine fois
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
	err := RedisClient.SAdd(ctx, permissionsKey, permissions).Err()
	if err != nil {
		return err
	}
	RedisClient.Expire(ctx, permissionsKey, 24*time.Hour)
	return nil
}

func RemoveCachedCdnPermissions(ctx context.Context, userID string) error {
	permissionsKey := fmt.Sprintf("cdn-permissions:%s", userID)
	err := RedisClient.Del(ctx, permissionsKey).Err()
	if err != nil {
		return err
	}
	log.Printf("Cache Redis supprimé pour %s", permissionsKey)
	return nil
}
