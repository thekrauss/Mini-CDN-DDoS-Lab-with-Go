package repositories

import (
	"time"

	"github.com/google/uuid"
)

// Structure de l'utilisateur
type Utilisateur struct {
	IDUtilisateur     uuid.UUID `json:"id_utilisateur" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	LoginID           string    `json:"login_id"`
	Nom               string    `json:"nom"`
	Prenom            string    `json:"prenom"`
	Email             string    `json:"email" gorm:"unique"`
	Genre             string    `json:"genre"`
	Telephone         string    `json:"telephone"`
	MotDePasse        string    `json:"mot_de_passe"`
	Role              string    `json:"role"`
	Permissions       string    `json:"permissions"`
	IDEcole           uuid.UUID `json:"id_ecole"`
	DateInscription   time.Time `json:"date_inscription"`
	DerniereConnexion time.Time `json:"derniere_connexion"`
	ClasseID          uuid.UUID `json:"classe_id,omitempty"`
	TokenExp          time.Time `json:"token_exp"`
	MFAEnabled        bool      `json:"mfa_enabled"`
	PhotoProfil       string    `json:"photo_profil"`
}
