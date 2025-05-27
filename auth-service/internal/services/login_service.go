package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/db"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/repositories"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
)

var ErrMFARequired = errors.New("MFA_REQUIRED")

type AuthService struct {
	Store *db.DBStore
}

func AuthenticateUser(db *sql.DB, identifier, password, ipAddress, userAgent string) (string, uuid.UUID, string, error) {
	var user repositories.Utilisateur
	var query string

	if strings.Contains(identifier, "@") {
		query = `
			SELECT id_utilisateur, mot_de_passe, role, email, telephone, mfa_enabled, prenom, nom, login_id
			FROM utilisateurs 
			WHERE email = $1
		`
	} else {
		query = `
			SELECT id_utilisateur, mot_de_passe, role, email, telephone, mfa_enabled, prenom, nom, login_id
			FROM utilisateurs 
			WHERE login_id = $1
		`
	}

	log.Printf("Executing query: %s with identifier: %s", query, identifier)

	err := db.QueryRow(query, identifier).Scan(
		&user.IDUtilisateur, &user.MotDePasseHash, &user.Role,
		&user.Email, &user.Telephone, &user.MFAEnabled,
		&user.Prenom, &user.Nom, &user.LoginID,
	)
	if err != nil {
		log.Printf("Database query error for user '%s': %v", identifier, err)
		if err == sql.ErrNoRows {
			return "", uuid.Nil, "", fmt.Errorf("incorrect identifier or password")
		}
		return "", uuid.Nil, "", fmt.Errorf("database error")
	}

	log.Printf("User found: ID=%s, Role=%s, MFA=%t, LoginID=%s", user.IDUtilisateur.String(), user.Role, user.MFAEnabled, user.LoginID)

	if !auth.CheckPasswordHash(password, user.MotDePasseHash) {
		logFailedAttempt(db, identifier, ipAddress, userAgent, user.Role)
		if isNewDevice(db, identifier, ipAddress, userAgent) {
			go SendSecurityAlerteEmail(user.Email, ipAddress, userAgent)
		}
		return "", uuid.Nil, "", fmt.Errorf("identifiant or incorrect password")
	}

	token, err := auth.GenerateJWT(user.IDUtilisateur, user.LoginID, user.Role)
	if err != nil {
		return "", uuid.Nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user.IDUtilisateur, user.Role, nil
}

func isNewDevice(db *sql.DB, identifier, ipAddress, userAgent string) bool {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM login_attempts 
		WHERE identifier = $1 
		  AND ip_address = $2 
		  AND user_agent = $3
		  AND successful = true
		  AND attempt_time > NOW() - INTERVAL '30 days'
	`
	err := db.QueryRow(query, identifier, ipAddress, userAgent).Scan(&count)
	if err != nil {
		log.Printf("Error checking device history: %v", err)
		return false
	}

	return count == 0
}

func logFailedAttempt(db *sql.DB, identifier, ipAddress, userAgent, role string) {
	var attemptCount int
	queryCheck := `
		SELECT COUNT(*) FROM login_attempts 
		WHERE identifier = $1 AND ip_address = $2 AND user_agent = $3 AND attempt_time > NOW() - INTERVAL '5 minutes'
	`
	err := db.QueryRow(queryCheck, identifier, ipAddress, userAgent).Scan(&attemptCount)
	if err != nil {
		log.Printf("Error checking failed attempts: %v", err)
		return
	}

	if attemptCount > 0 {
		return
	}

	queryInsert := `
		INSERT INTO login_attempts (identifier, ip_address, user_agent, successful, attempt_time, role)
		VALUES ($1, $2, $3, $4, NOW(), $5)
	`
	_, err = db.Exec(queryInsert, identifier, ipAddress, userAgent, false, role)
	if err != nil {
		log.Printf("Failed to log failed attempt for %s: %v", identifier, err)
	}
}
