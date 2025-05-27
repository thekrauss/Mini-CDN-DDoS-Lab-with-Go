package services

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
)

func GenerateLoginID(prenom, nom string) string {
	prenomParts := strings.Fields(prenom)
	nomParts := strings.Fields(nom)

	if len(prenomParts) == 0 || len(nomParts) == 0 {
		log.Println("Erreur: prénom ou nom vide")
		return ""
	}

	initialePrenom := strings.ToLower(string(prenomParts[0][0]))
	premierNom := strings.ToLower(nomParts[0])

	return initialePrenom + premierNom // Ex: "Marie Claire", "Ngoma Mbemba" → "mngoma"
}

// un mot de passe sécurisé avec la longueur spécifiée
func GeneratePassword(length int) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("La longueur du mot de passe doit être d'au moins 8 caractères")
	}

	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	digits := "0123456789"
	special := "!@#$%^&*()-_=+[]{}|;:,.<>?/"

	allChars := uppercase + lowercase + digits + special
	password := make([]rune, length)

	password[0] = rune(uppercase[getRandomIndex(len(uppercase))])
	password[1] = rune(lowercase[getRandomIndex(len(lowercase))])
	password[2] = rune(digits[getRandomIndex(len(digits))])
	password[3] = rune(special[getRandomIndex(len(special))])

	for i := 4; i < length; i++ {
		password[i] = rune(allChars[getRandomIndex(len(allChars))])
	}

	shuffle(password)

	return string(password), nil
}

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

// un index aléatoire sécurisé
func getRandomIndex(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}
	return int(n.Int64())
}

func shuffle(password []rune) {
	for i := range password {
		j := getRandomIndex(len(password))
		password[i], password[j] = password[j], password[i]
	}
}
