package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/repositories"
)

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*repositories.Utilisateur, error) {

	cachedKey := fmt.Sprintf("user:%s,", userID)

	cached, err := RedisClient.Get(ctx, cachedKey).Result()
	if err != nil {
		var user repositories.Utilisateur
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}
	var user repositories.Utilisateur
	query := `
		SELECT id_utilisateur, nom, prenom, email, telephone, role, permissions, mfa_enabled, photo_profil
		FROM utilisateurs
		WHERE id_utilisateur = $1
	`

	err = s.Store.DB.QueryRow(query, userID).Scan(
		&user.IDUtilisateur, &user.Nom, &user.Prenom, &user.Email,
		&user.Telephone, &user.Role, &user.Permissions, &user.MFAEnabled, &user.PhotoProfil,
	)
	if err != nil {
		return &user, err
	}

	if raw, err := json.Marshal(&user); err == nil {
		if err := RedisClient.Set(ctx, cachedKey, raw, 60*time.Minute).Err(); err != nil {
			log.Println("Failed to set Redis cache", err)
		}
	}

	return &user, nil
}

func (s *AuthService) GetUserEmailByID(ctx context.Context, userID uuid.UUID) (*string, error) {
	cachedKey := fmt.Sprintf("user:%s,", userID)
	cached, err := RedisClient.Get(ctx, cachedKey).Result()
	if err != nil {
		var email string
		if err := json.Unmarshal([]byte(cached), &email); err == nil {
			return &email, nil
		}
	}

	var email string
	query := "SELECT email FROM utilisateurs WHERE id_utilisateur = $1"
	err = s.Store.DB.QueryRow(query, userID).Scan(&email)
	if err != nil {
		return &email, err
	}

	if raw, err := json.Marshal(&email); err != nil {
		if err := RedisClient.Set(ctx, cachedKey, raw, 60*time.Minute).Err(); err != nil {
			log.Println("Failed to set Redis cache", err)
		}
	}

	return &email, nil
}

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*repositories.Utilisateur, error) {

	cachedKey := fmt.Sprintf("user:%s,", email)
	cached, err := RedisClient.Get(ctx, cachedKey).Result()
	if err != nil {
		var user repositories.Utilisateur
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	var user repositories.Utilisateur
	query := `SELECT id_utilisateur, nom, prenom, email, genre, telephone, mot_de_passe, role, photo_profil 
			  FROM utilisateurs 
			  WHERE email = $1`

	err = s.Store.DB.QueryRow(query, email).Scan(
		&user.IDUtilisateur, &user.Nom, &user.Prenom, &user.Email, &user.Genre,
		&user.Telephone, &user.MotDePasseHash, &user.Role, &user.PhotoProfil,
	)

	if raw, err := json.Marshal(&user); err == nil {
		if err := RedisClient.Set(ctx, cachedKey, raw, 60*time.Minute).Err(); err != nil {
			log.Println("Failed to set Redis cache", err)
		}
	}

	return &user, err
}

func (s *AuthService) SaveRefreshToken(userID uuid.UUID, token string) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '7 days')
		ON CONFLICT (user_id) DO UPDATE
		SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at, created_at = NOW();
	`

	_, err := s.Store.DB.Exec(query, userID, token)
	if err != nil {
		log.Printf("Error saving refresh token for user %s: %v", userID, err)
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	log.Printf("Refresh token saved for user %s", userID)
	return nil
}
