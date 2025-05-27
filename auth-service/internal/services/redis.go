package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/repositories"
)

var RedisClient *redis.Client

func InitRedis(cfg *config.Config) {

	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf(" Impossible de se connecter à Redis: %v", err)
	} else {
		log.Printf(" Connexion Redis réussie: %s", pong)
	}
}

func StoreUserSessionInRedis(ctx context.Context, user *repositories.Utilisateur, ipAddress, userAgent string) error {
	redisKey := fmt.Sprintf("user:session:%s", user.IDUtilisateur.String())

	sessionData := map[string]any{
		"id_utilisateur":     user.IDUtilisateur.String(),
		"nom":                user.Nom,
		"prenom":             user.Prenom,
		"email":              user.Email,
		"telephone":          user.Telephone,
		"role":               user.Role,
		"permissions":        user.Permissions,
		"id_ecole":           user.IDEcole.String(),
		"mfa_enabled":        user.MFAEnabled,
		"photo_profil":       user.PhotoProfil,
		"adresse_ip":         ipAddress,
		"user_agent":         userAgent,
		"derniere_connexion": time.Now().Format(time.RFC3339),
	}

	err := RedisClient.HSet(ctx, redisKey, sessionData).Err()
	if err != nil {
		log.Printf("Erreur lors de l'enregistrement de la session utilisateur en Redis: %v", err)
		return err
	}

	// Expiration de la session en 24h
	RedisClient.Expire(ctx, redisKey, 24*time.Hour)

	log.Printf("Session utilisateur enregistrée en Redis [ID: %s]", user.IDUtilisateur.String())
	return nil
}
