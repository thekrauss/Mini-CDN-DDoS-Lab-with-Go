package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/services"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

// register gRPC handler
func (s *AuthServer) CreateAdmin(ctx context.Context, req *pb.CreateAdminRequest) (*pb.CreateAdminResponse, error) {

	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}

	return s.Service.CreateAdmin(ctx, req, claims)
}

func (s *AuthServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	log.Printf("Mise à jour de l'utilisateur %s", req.UtilisateurId)

	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}
	//  les permissions
	if err := s.Service.CheckAdminPermissions(ctx, claims, "MANAGE_ADMIN"); err != nil {
		return nil, err
	}

	//  si l'utilisateur existe
	var exists bool
	queryCheck := `SELECT EXISTS (SELECT 1 FROM utilisateurs WHERE id_utilisateur = $1)`
	err = s.Store.DB.QueryRowContext(ctx, queryCheck, req.UtilisateurId).Scan(&exists)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur lors de la vérification de l'utilisateur")
	}
	if !exists {
		return nil, status.Errorf(codes.NotFound, "L'utilisateur avec l'ID %s n'existe pas", req.UtilisateurId)
	}

	//  si l'email ou le téléphone existent déjà pour un autre utilisateur
	var existingUserID string
	checkQuery := `SELECT id_utilisateur FROM utilisateurs WHERE (email = $1 OR telephone = $2) AND id_utilisateur != $3`
	err = s.Store.DB.QueryRowContext(ctx, checkQuery, req.Email, req.Telephone, req.UtilisateurId).Scan(&existingUserID)
	if err != nil && err != sql.ErrNoRows {
		return nil, status.Errorf(codes.Internal, "Erreur lors de la vérification de l'email ou du téléphone")
	}
	if existingUserID != "" {
		return nil, status.Errorf(codes.AlreadyExists, "Email ou téléphone déjà utilisé par un autre utilisateur")
	}

	// met à jour des informations de l'utilisateur, y compris la photo de profil
	queryUpdate := `
		UPDATE utilisateurs 
		SET nom = $1, prenom = $2, email = $3, telephone = $4, role = $5, date_mise_a_jour = NOW()
		WHERE id_utilisateur = $7
	`
	_, err = s.Store.DB.ExecContext(ctx, queryUpdate,
		req.Nom, req.Prenom, req.Email, req.Telephone, req.Role, req.UtilisateurId)

	if err != nil {
		log.Printf("Erreur SQL lors de la mise à jour : %v", err)
		return nil, status.Errorf(codes.Internal, "Échec de la mise à jour de l'utilisateur")
	}

	// Supprime le cache Redis de l'utilisateur
	redisKey := fmt.Sprintf("user_info:%s", req.UtilisateurId)
	err = services.RedisClient.Del(ctx, redisKey).Err()
	if err != nil {
		log.Printf("Impossible de supprimer le cache Redis : %v", err)
	}

	log.Printf("Mise à jour réussie pour l'utilisateur %s", req.UtilisateurId)
	return &pb.UpdateUserResponse{
		Message: fmt.Sprintf("L'utilisateur %s a été mis à jour avec succès", req.UtilisateurId),
	}, nil
}

// Handler pour supprimer un utilisateur
func (s *AuthServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log.Printf("suppression de l'utilisateur : %s", req.UtilisateurId)

	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}
	//  les permissions
	if err := s.Service.CheckAdminPermissions(ctx, claims, "MANAGE_ADMIN"); err != nil {
		return nil, err
	}

	var exists bool
	err = s.Store.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM utilisateurs WHERE id_utilisateur = $1)", req.UtilisateurId).Scan(&exists)
	if err != nil {
		log.Printf("Erreur SQL : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la vérification de l'utilisateur")
	}

	if !exists {
		return &pb.DeleteUserResponse{
			Success: false,
			Message: "L'utilisateur n'existe pas",
		}, status.Errorf(codes.NotFound, "L'utilisateur avec l'ID %s n'existe pas", req.UtilisateurId)
	}

	// suppression de l'utilisateur
	_, err = s.Store.DB.Exec("DELETE FROM utilisateurs WHERE id_utilisateur = $1", req.UtilisateurId)
	if err != nil {
		log.Printf("Erreur SQL lors de la suppression : %v", err)
		return nil, status.Errorf(codes.Internal, "Échec de la suppression de l'utilisateur")
	}

	log.Printf("Utilisateur %s supprimé avec succès", req.UtilisateurId)
	return &pb.DeleteUserResponse{
		Success: true,
		Message: "Utilisateur supprimé avec succès",
	}, nil
}

func (s *AuthServer) ListAllAdmins(ctx context.Context, req *pb.ListAllAdminsRequest) (*pb.ListAllAdminsResponse, error) {
	log.Println("🔍 Récupération des administrateurs avec filtres...")

	cacheKey := fmt.Sprintf("all_admins:limit=%d:offset=%d:query=%s:tenant=%s:active=%t", req.Limit, req.Offset, req.Query, req.TenantId, req.IsActive)

	if cachedData, err := services.RedisClient.Get(ctx, cacheKey).Bytes(); err == nil {
		var cachedAdmins pb.ListAllAdminsResponse
		if err := json.Unmarshal(cachedData, &cachedAdmins); err == nil {
			log.Println("Liste des administrateurs récupérée depuis Redis")
			return &cachedAdmins, nil
		}
	}

	// Base de la requête SQL
	baseQuery := `
		FROM utilisateurs
		WHERE role ILIKE '%admin%' 
		  AND (nom ILIKE $1 OR prenom ILIKE $1 OR email ILIKE $1)
		  AND tenant_id = $2
		  AND is_active = $3
	`

	query := `SELECT id_utilisateur, nom, prenom, email, telephone, role ` + baseQuery + `
		ORDER BY date_inscription DESC
		LIMIT $4 OFFSET $5`

	countQuery := `SELECT COUNT(*) ` + baseQuery

	rows, err := s.Store.DB.QueryContext(ctx, query, "%"+req.Query+"%", req.TenantId, req.IsActive, req.Limit, req.Offset)
	if err != nil {
		log.Printf(" Erreur SQL récupération admins : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur SQL")
	}
	defer rows.Close()

	var admins []*pb.GetAdminInfoResponse
	for rows.Next() {
		var admin pb.GetAdminInfoResponse
		if err := rows.Scan(&admin.UtilisateurId, &admin.Nom, &admin.Prenom, &admin.Email, &admin.Telephone, &admin.Role); err != nil {
			return nil, status.Errorf(codes.Internal, "Erreur lecture ligne admin")
		}
		admins = append(admins, &admin)
	}

	var total int32
	if err := s.Store.DB.QueryRowContext(ctx, countQuery, "%"+req.Query+"%", req.TenantId, req.IsActive).Scan(&total); err != nil {
		log.Printf(" Erreur comptage admins : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur SQL count")
	}

	adminsResponse := &pb.ListAllAdminsResponse{
		Admins: admins,
		Total:  total,
	}
	if jsonData, err := json.Marshal(adminsResponse); err == nil {
		services.RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute)
	}

	return adminsResponse, nil
}

func (s *AuthServer) GetAdminByID(ctx context.Context, req *pb.GetAdminByIDRequest) (*pb.GetAdminInfoResponse, error) {
	log.Printf("Recherche d’un administrateur : %s", req.UtilisateurId)

	var admin pb.GetAdminInfoResponse
	query := `
		SELECT id_utilisateur, nom, prenom, email, telephone, role
		FROM utilisateurs
		WHERE id_utilisateur = $1 AND role ILIKE '%admin%'
	`

	err := s.Store.DB.QueryRowContext(ctx, query, req.UtilisateurId).
		Scan(&admin.UtilisateurId, &admin.Nom, &admin.Prenom, &admin.Email, &admin.Telephone, &admin.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "Admin %s introuvable", req.UtilisateurId)
		}
		log.Printf(" Erreur SQL : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur SQL")
	}

	return &admin, nil
}
