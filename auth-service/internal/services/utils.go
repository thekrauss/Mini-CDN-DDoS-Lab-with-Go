package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
)

func UpdateLastActivity(db *sql.DB, userID uuid.UUID) {
	query := `UPDATE utilisateurs SET last_activity = NOW() WHERE id_utilisateur = $1`
	_, err := db.Exec(query, userID)
	if err != nil {
		log.Printf("Failed to update last activity for user %s: %v", userID, err)
	}
}

func MaskSensitiveData(input string) string {
	if len(input) <= 3 {
		return "***"
	}
	return input[:3] + strings.Repeat("*", len(input)-3)
}

func EmailExists(db *sql.DB, email string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM utilisateurs WHERE email = $1"
	err := db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("database query error: %w", err)
	}
	return count > 0, nil
}

func ParseRole(roleStr string) (pb.Role, error) {
	if !config.IsRoleA(roleStr) && !config.IsRoleB(roleStr) {
		return pb.Role(0), fmt.Errorf("rôle inconnu ou non autorisé : %s", roleStr)
	}

	roleEnum, ok := pb.Role_value[roleStr]
	if !ok {
		return pb.Role(0), fmt.Errorf("rôle %s non trouvé dans l'énumération gRPC", roleStr)
	}

	return pb.Role(roleEnum), nil
}
