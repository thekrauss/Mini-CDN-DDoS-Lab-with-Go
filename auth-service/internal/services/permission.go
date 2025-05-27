package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

func (s *AuthService) SetCdnPermissions(ctx context.Context, req *pb.SetCdnPermissionsRequest) (*pb.SetCdnPermissionsResponse, error) {
	log.Printf("Mise à jour des permissions Cdn pour l'utilisateur %s", req.UtilisateurId)

	// Vérifie localement si l'utilisateur existe
	var exists bool
	err := s.Store.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM utilisateurs WHERE id_utilisateur = $1)`, req.UtilisateurId).Scan(&exists)
	if err != nil {
		log.Printf("Erreur SQL lors de la vérification de l'utilisateur : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la vérification de l'utilisateur")
	}
	if !exists {
		return nil, status.Errorf(codes.NotFound, "L'utilisateur avec l'ID %s n'existe pas", req.UtilisateurId)
	}

	// Supprime les permissions existantes
	_, err = s.Store.DB.ExecContext(ctx, `DELETE FROM admin_permissions WHERE id_utilisateur = $1`, req.UtilisateurId)
	if err != nil {
		log.Printf("Erreur SQL lors de la suppression des anciennes permissions : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la mise à jour des permissions")
	}

	// Ajoute les nouvelles permissions
	for _, permission := range req.Permissions {
		_, err := s.Store.DB.ExecContext(ctx,
			`INSERT INTO admin_permissions (id_utilisateur, permission) VALUES ($1, $2)`,
			req.UtilisateurId, permission)
		if err != nil {
			log.Printf("Erreur SQL lors de l'ajout de la permission %s : %v", permission, err)
			return nil, status.Errorf(codes.Internal, "Erreur lors de l'ajout des permissions")
		}
	}

	// Mise à jour du cache Redis
	_ = RemoveCachedCdnPermissions(ctx, req.UtilisateurId)
	if err := CachCdnPermissionsInRedis(ctx, req.UtilisateurId, req.Permissions); err != nil {
		log.Printf("Impossible de mettre en cache Redis : %v", err)
	}

	log.Printf(" Permissions mises à jour avec succès pour %s", req.UtilisateurId)

	return &pb.SetCdnPermissionsResponse{
		Message: fmt.Sprintf("Permissions SYK mises à jour pour l'utilisateur %s", req.UtilisateurId),
	}, nil
}

func (s *AuthService) CheckAdminPermissions(ctx context.Context, claims *auth.Claims, permission string) error {

	adminUser, err := GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		log.Printf("Échec récupération infos admin depuis Redis: %v", err)
		return status.Errorf(codes.Internal, "Impossible de récupérer les informations de l'administrateur.")
	}

	// if l'utilisateur a un Role A, il a accès à tout
	if config.IsRoleA(adminUser.Role) {
		return nil
	}

	// if l'utilisateur un Role B, il doit avoir la permission spécifique
	if config.IsRoleB(adminUser.Role) {
		hasPerm, err := s.HasSykPermission(ctx, adminUser.IDUtilisateur, permission)
		if err != nil {
			log.Printf("Erreur lors de la vérification des permissions: %v", err)
			return status.Errorf(codes.Internal, "Erreur interne lors de la vérification des permissions")
		}
		if !hasPerm {
			return status.Errorf(codes.PermissionDenied, "Vous n'avez pas la permission d'effectuer cette action")
		}
		return nil
	}

	// if l'utilisateur n'est ni A ni B, accès refusé
	return status.Errorf(codes.PermissionDenied, "Accès refusé")
}

func (s *AuthService) HasSykPermission(ctx context.Context, userID, requiredPermission string) (bool, error) {
	permissionsKey := fmt.Sprintf("cdn-permissions:%s", userID)

	// verifi d'abord dans Redis
	exists, err := RedisClient.SIsMember(ctx, permissionsKey, requiredPermission).Result()
	if err == nil && exists {
		log.Printf("Permission '%s' trouvée en cache Redis pour %s", requiredPermission, userID)
		return true, nil
	}

	//  dans PostgreSQL si la permission n'est pas en cache
	var count int
	err = s.Store.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_permissions WHERE id_utilisateur = $1 AND permission = $2`,
		userID, requiredPermission).Scan(&count)
	if err != nil {
		return false, status.Errorf(codes.Internal, "Erreur lors de la vérification de la permission")
	}

	// Si permission trouvée en db, met en cache pour la prochaine fois
	if count > 0 {
		err = CachCdnPermissionsInRedis(ctx, userID, []string{requiredPermission})
		if err != nil {
			log.Printf("Impossible de mettre en cache Redis : %v", err)
		}
		return true, nil
	}

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
