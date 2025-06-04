package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/repositories"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AuthService) CreateAdmin(ctx context.Context, req *pb.CreateAdminRequest, claims *auth.Claims) (*pb.CreateAdminResponse, error) {

	if req.IdTenant == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id_tenant requis pour créer un administrateur")
	}

	tenantID, err := uuid.Parse(req.IdTenant)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "id_tenant invalide : %v", err)
	}

	var exists bool
	err = s.Store.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM tenant WHERE id_tenant = $1)`, tenantID).Scan(&exists)
	if err != nil || !exists {
		return nil, status.Errorf(codes.NotFound, "Le tenant %s est introuvable", req.IdTenant)
	}

	adminUser, err := GetUserInfoFromRedis(ctx, claims.UserID.String())
	if err != nil {
		log.Printf("Échec récupération infos admin depuis Redis: %v", err)
		return nil, status.Errorf(codes.Internal, "Impossible de récupérer les informations de l'administrateur.")
	}

	if err := s.CheckAdminPermissions(ctx, claims, "MANAGE_ADMIN"); err != nil {
		return nil, err
	}

	// identifiant de connexion
	loginID := GenerateLoginID(req.Prenom, req.Nom)

	adminData := &repositories.Utilisateur{
		IDUtilisateur:   uuid.New(),
		Nom:             req.Nom,
		Prenom:          req.Prenom,
		Email:           req.Email,
		Genre:           req.Genre,
		Telephone:       req.Telephone,
		Role:            req.Role,
		DateInscription: time.Now(),
		LoginID:         loginID,
		TenantID:        tenantID,
		PhotoProfil:     "https://storage.googleapis.com/mon-bucket/profil-default.jpg",
	}

	// crée l'utilisateur ou récupère un existant
	userID, tempPassword, err := s.ensureOrCreateAdmin(ctx, adminData)
	if err != nil {
		s.RollbackAdminCreation(ctx, userID)
		log.Printf("Erreur lors de la création de l'administrateur : %v", err)
		return nil, status.Errorf(codes.Internal, "Impossible de créer l'administrateur.")
	}

	var count int
	if err := s.Store.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM utilisateurs WHERE id_utilisateur = $1
	`, adminData.IDUtilisateur).Scan(&count); err != nil {
		log.Printf("Erreur de vérification de l'existence : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la vérification de l'existence.")
	}
	if count > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "L'utilisateur %s est déjà administrateur.", userID)
	}

	// ajout des permissions si nécessaires
	if len(req.Permissions) > 0 {
		for _, permission := range req.Permissions {
			_, err := s.Store.DB.ExecContext(ctx, `
				INSERT INTO utilisateurs_permissions (id_utilisateur, permission)
				VALUES ($1, $2)
			`, adminData.IDUtilisateur, permission)
			if err != nil {
				s.RollbackAdminCreation(ctx, userID)
				log.Printf(" Erreur insertion permission %s : %v", permission, err)
				return nil, status.Errorf(codes.Internal, "Erreur lors de l'ajout des permissions.")
			}
		}

		if err := CachCdnPermissionsInRedis(ctx, userID, req.Permissions); err != nil {
			log.Printf("Erreur cache permissions : %v", err)
		}
	}

	// audit log
	ip, userAgent := GetRequestMetadata(ctx)
	_ = s.LogAction(ctx, AuditLog{
		AdminID:    claims.UserID,
		Role:       adminUser.Role,
		Action:     "Création administrateur",
		TargetID:   adminData.IDUtilisateur,
		TargetType: "Administrateur",
		Details:    fmt.Sprintf("L'utilisateur %s a été assigné comme administrateur.", adminData.Email),
		IPAddress:  ip,
		UserAgent:  userAgent,
		ActionTime: time.Now(),
		Status:     "succès",
	})

	// envoi email
	go s.SendAdminSecurityAlertEmail(
		adminUser.Email,
		adminData.Email,
		adminData.Nom,
		adminData.Prenom,
		adminData.LoginID,
		tempPassword,
		"Administration CDN",
		ip,
		userAgent,
	)

	log.Printf("Administrateur %s créé avec succès avec pour mot de passe:  %s.", adminData.IDUtilisateur, tempPassword)
	return &pb.CreateAdminResponse{
		Message: fmt.Sprintf("Administrateur %s créé avec succès.", adminData.IDUtilisateur),
	}, nil
}

func (s *AuthService) ensureOrCreateAdmin(ctx context.Context, data *repositories.Utilisateur) (string, string, error) {
	exists, err := EmailExists(s.Store.DB, data.Email)
	if err != nil {
		return "", "", fmt.Errorf("erreur de vérification de l'existence de l'utilisateur : %w", err)
	}

	if exists {
		user, err := s.GetUserByEmail(ctx, data.Email)
		if err != nil {
			return "", "", fmt.Errorf("utilisateur existant non récupérable : %w", err)
		}

		return user.IDUtilisateur.String(), "", nil
	}

	password, err := GeneratePassword(12)
	if err != nil {
		return "", "", fmt.Errorf("échec génération mot de passe : %w", err)
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return "", "", fmt.Errorf("échec hashage mot de passe : %w", err)
	}

	data.MotDePasseHash = hashedPassword

	query := `
		INSERT INTO utilisateurs (
			id_utilisateur, login_id, nom, prenom, email, genre,
			telephone, mot_de_passe_hash, role, date_inscription, photo_profil
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`

	_, err = s.Store.DB.ExecContext(ctx, query,
		data.IDUtilisateur, data.LoginID, data.Nom, data.Prenom,
		data.Email, data.Genre, data.Telephone,
		data.MotDePasseHash, data.Role, data.DateInscription, data.PhotoProfil)

	if err != nil {
		return "", "", fmt.Errorf("échec insertion base : %w", err)
	}

	return data.IDUtilisateur.String(), password, nil
}

// supprime un utilisateur localement en cas d’échec
func (s *AuthService) RollbackAdminCreation(ctx context.Context, userID string) {
	query := `DELETE FROM utilisateurs WHERE id_utilisateur = $1`

	_, err := s.Store.DB.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("Impossible de supprimer l'utilisateur %s après rollback: %v", userID, err)
	} else {
		log.Printf("Rollback réussi : utilisateur %s supprimé", userID)
	}
}
