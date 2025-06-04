package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AuthService) CreateTenant(ctx context.Context, req *pb.CreateTenantRequest, claims *auth.Claims) (*pb.CreateTenantResponse, error) {
	log.Println("Requête reçue : CreateTenant")

	adminUser, err := GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		log.Printf("Erreur récupération infos admin depuis Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "Impossible de récupérer les informations de l'administrateur")
	}

	if req.Nom == "" || req.Adresse == "" || req.Ville == "" || req.CodePostal == "" ||
		req.ContactTelephone == "" || req.ContactEmail == "" || req.DirecteurNom == "" ||
		req.DirecteurContact == "" || req.TypeEtablissement == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Tous les champs obligatoires doivent être remplis")
	}

	ip, userAgent := GetRequestMetadata(ctx)

	idTenant := uuid.New()

	logoURL := req.LogoUrl
	if logoURL == "" {
		logoURL = "https://storage.googleapis.com/mon-bucket/profil-default.jpg"
	}

	query := `
		INSERT INTO tenant (
			id_tenant, nom, adresse, ville, code_postal,
			contact_telephone, contact_email, directeur_nom, directeur_contact,
			type_etablissement, parametres_specifiques, date_creation, validation_status, logo_url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,
		        $10, $11, $12, $13, $14)
	`

	_, err = s.Store.DB.ExecContext(ctx, query,
		idTenant, req.Nom, req.Adresse, req.Ville, req.CodePostal,
		req.ContactTelephone, req.ContactEmail, req.DirecteurNom, req.DirecteurContact,
		req.TypeEtablissement, req.ParametresSpecifiques, time.Now(), "En attente", logoURL,
	)

	if err != nil {
		log.Printf("Erreur insertion SQL : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la création du tenant")
	}

	_ = s.LogAction(ctx, AuditLog{
		AdminID:    claims.UserID,
		Role:       adminUser.Role,
		Action:     "Création de tenant",
		TargetID:   idTenant,
		TargetType: "Tenant",
		Details:    fmt.Sprintf("Création du tenant %s", req.Nom),
		ActionTime: time.Now(),
		Status:     "En attente",
		IPAddress:  ip,
		UserAgent:  userAgent,
	})

	return &pb.CreateTenantResponse{
		Message:  "Tenant créé avec succès",
		IdTenant: idTenant.String(),
	}, nil
}

func (s *AuthService) DeleteTenant(ctx context.Context, req *pb.DeleteTenantRequest, claims *auth.Claims) (*emptypb.Empty, error) {
	log.Printf("Suppression du tenant: %s", req.IdTenant)

	adminUser, err := GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		log.Printf("Erreur récupération admin depuis Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "Impossible de récupérer les informations de l'administrateur")
	}

	if err := s.CheckAdminPermissions(ctx, claims, "DELETE_SCHOOL"); err != nil {
		return nil, err
	}

	ip, userAgent := GetRequestMetadata(ctx)

	var nomTenant string
	checkQuery := `SELECT nom FROM tenant WHERE id_tenant = $1`
	err = s.Store.DB.QueryRow(checkQuery, req.IdTenant).Scan(&nomTenant)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "Le tenant avec l'ID %s n'existe pas", req.IdTenant)
		}
		log.Printf("Erreur lors de la récupération du nom du tenant: %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur interne")
	}

	deleteQuery := `DELETE FROM tenant WHERE id_tenant = $1`
	res, err := s.Store.DB.ExecContext(ctx, deleteQuery, req.IdTenant)
	if err != nil {
		log.Printf("Erreur SQL lors de la suppression: %v", err)
		return nil, status.Errorf(codes.Internal, "Échec de la suppression du tenant")
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, status.Errorf(codes.NotFound, "Aucun tenant supprimé, vérifiez l'ID")
	}

	logEntry := AuditLog{
		AdminID:    claims.UserID,
		Role:       adminUser.Role,
		Action:     "Suppression de tenant",
		TargetID:   uuid.MustParse(req.IdTenant),
		TargetType: "Tenant",
		Details:    fmt.Sprintf("Suppression du tenant: %s (ID: %s), Raison: %s", nomTenant, req.IdTenant, req.Raison),
		IPAddress:  ip,
		UserAgent:  userAgent,
		ActionTime: time.Now(),
		Status:     "Succès",
	}

	if err := s.LogAction(ctx, logEntry); err != nil {
		log.Printf("Impossible d'enregistrer l'audit: %v", err)
	}

	_ = RedisClient.Del(ctx, fmt.Sprintf("tenant_admins:%s", req.IdTenant)).Err()
	_ = RedisClient.Del(ctx, fmt.Sprintf("admin_info:%s", req.UtilisateurId)).Err()

	return &emptypb.Empty{}, nil
}
