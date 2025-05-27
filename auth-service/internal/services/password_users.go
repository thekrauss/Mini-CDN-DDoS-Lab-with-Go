package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/google/uuid"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	Store *sql.DB
}

func (s *AuthService) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {

	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "L'email est requis")
	}

	if !isValidEmail(req.Email) {
		return nil, status.Errorf(codes.InvalidArgument, "Format d'email invalide")
	}

	var userID string
	query := "SELECT id_utilisateur FROM utilisateurs WHERE email = $1"
	err := s.Store.QueryRowContext(ctx, query, req.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.ForgotPasswordResponse{
				Message: "Si votre email est valide, un lien de réinitialisation vous sera envoyé.",
			}, nil
		}
		log.Printf("Erreur lors de la vérification de l'email %s: %v", req.Email, err)
		return nil, status.Errorf(codes.Internal, "Erreur interne lors de la vérification de l'email")
	}

	ctxRedis := context.Background()
	requestsKey := fmt.Sprintf("password_reset_attempts:%s", req.Email)
	count, _ := RedisClient.Get(ctxRedis, requestsKey).Int()
	if count >= 3 {
		log.Printf("Trop de demandes de réinitialisation pour %s", req.Email)
		return nil, status.Errorf(codes.ResourceExhausted, "Trop de tentatives, veuillez attendre avant de réessayer.")
	}

	resetToken := uuid.New().String()
	tokenKey := fmt.Sprintf("password_reset_token:%s", req.Email)

	err = RedisClient.Set(ctxRedis, tokenKey, resetToken, 15*time.Minute).Err()
	if err != nil {
		log.Printf("Erreur stockage token dans Redis : %v", err)
		return nil, status.Errorf(codes.Internal, "Impossible de générer le lien de réinitialisation.")
	}

	RedisClient.Incr(ctxRedis, requestsKey)
	RedisClient.Expire(ctxRedis, requestsKey, 10*time.Minute)

	resetLink, err := GenerateResetLinkREST(ctx, req.Email)
	if err != nil {
		log.Printf("Erreur lors de la génération du lien Firebase pour %s: %v", req.Email, err)
		return nil, status.Errorf(codes.Internal, "Impossible de générer le lien de réinitialisation.")
	}

	subjectMsg := " Réinitialisation de votre mot de passe"
	bodyMsg := fmt.Sprintf(
		"Bonjour,\n\nPour réinitialiser votre mot de passe, cliquez sur le lien suivant:\n\n%s\n\nCe lien expire dans 5 minutes.\n\nCordialement,\n",
		resetLink,
	)

	if err := sendEmail(req.Email, subjectMsg, bodyMsg); err != nil {
		log.Printf("Erreur d'envoi du lien à %s: %v", req.Email, err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de l'envoi de l'email.")
	}

	log.Printf("Lien de réinitialisation envoyé à %s", req.Email)
	return &pb.ForgotPasswordResponse{
		Message: "Si votre email est valide, un lien de réinitialisation vous a été envoyé.",
	}, nil
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
