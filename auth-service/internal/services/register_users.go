package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/repositories"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
)

func RegisterUser(db *sql.DB, req *pb.RegisterRequest) (uuid.UUID, string, error) {

	// Génération de l'ID de connexion (`login_id`)
	loginID := GenerateLoginID(req.Prenom, req.Nom)

	newUser := repositories.Utilisateur{
		IDUtilisateur:   uuid.New(),
		Nom:             req.Nom,
		Prenom:          req.Prenom,
		Email:           req.Email,
		Genre:           req.Genre,
		Telephone:       req.Telephone,
		MotDePasse:      req.MotDePasse,
		Role:            req.Role,
		IDEcole:         uuid.MustParse(req.IdEcole),
		DateInscription: time.Now(),
		LoginID:         loginID,
	}

	query := `INSERT INTO utilisateurs (id_utilisateur, nom, prenom, email, genre, telephone, mot_de_passe, role, id_ecole, date_inscription, login_id, photo_profil) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := db.Exec(query, newUser.IDUtilisateur, newUser.Nom, newUser.Prenom, newUser.Email, newUser.Genre,
		newUser.Telephone, newUser.MotDePasse, newUser.Role, newUser.IDEcole, newUser.DateInscription, newUser.LoginID, newUser.PhotoProfil)

	if err != nil {
		log.Printf("Database error: %v", err)
		return uuid.Nil, "", fmt.Errorf("failed to register user")
	}

	return newUser.IDUtilisateur, newUser.LoginID, nil
}

func GenerateLoginID(prenom, nom string) string {

	prenomParts := strings.Fields(prenom)
	nomParts := strings.Fields(nom)

	if len(prenomParts) == 0 || len(nomParts) == 0 {
		log.Println("Erreur: prénom ou nom vide")
		return ""
	}

	initialePrenom := strings.ToLower(string(prenomParts[0][0]))
	premierNom := strings.ToLower(nomParts[0])

	loginID := initialePrenom + premierNom
	return loginID //"Marie Claire", "Ngoma Mbemba" --> Résultat : mngoma
}
