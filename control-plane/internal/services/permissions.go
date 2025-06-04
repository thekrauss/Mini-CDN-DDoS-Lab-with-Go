package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/db"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeService struct {
	Store *db.DBStore
}

func (s *NodeService) CheckAdminPermissions(ctx context.Context, claims *auth.Claims, permission string) error {

	adminUser, err := GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		return status.Errorf(codes.Internal, "Impossible de récupérer les informations de l'administrateur.")
	}

	// Cas classiques
	if config.IsRoleA(adminUser.Role) {
		return nil
	}

	if config.IsRoleB(adminUser.Role) {
		hasPerm, err := s.HasPermission(ctx, adminUser.IDUtilisateur, permission)
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

func (s *NodeService) HasPermission(ctx context.Context, userID, requiredPermission string) (bool, error) {
	permissionsKey := fmt.Sprintf("cdn-permissions:%s", userID)

	// verifi d'abord dans Redis
	exists, err := RedisClient.SIsMember(ctx, permissionsKey, requiredPermission).Result()
	if err == nil && exists {
		log.Printf("Permission '%s' trouvée en cache Redis pour %s", requiredPermission, userID)
		return true, nil
	}

	//  dans PostgreSQL si la permission n'est pas en cache
	// var count int
	// err = s.Store.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM utilisateurs_permissions WHERE id_utilisateur = $1 AND permission = $2`,
	// 	userID, requiredPermission).Scan(&count)
	// if err != nil {
	// 	return false, status.Errorf(codes.Internal, "Erreur lors de la vérification de la permission")
	// }

	// Si permission trouvée en db, met en cache pour la prochaine fois
	// if count > 0 {
	// 	err = CachCdnPermissionsInRedis(ctx, userID, []string{requiredPermission})
	// 	if err != nil {
	// 		log.Printf("Impossible de mettre en cache Redis : %v", err)
	// 	}
	// 	return true, nil
	// }

	log.Printf("Permission '%s' non trouvée pour %s", requiredPermission, userID)
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
