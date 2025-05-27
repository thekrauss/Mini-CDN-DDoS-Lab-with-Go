package repositories

import (
	"time"

	"github.com/google/uuid"
)

type Utilisateur struct {
	IDUtilisateur     uuid.UUID `json:"id_utilisateur" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"` // Identifiant unique
	LoginID           string    `json:"login_id"`                                                              // Identifiant de connexion généré
	Nom               string    `json:"nom"`                                                                   // Nom de l'utilisateur
	Prenom            string    `json:"prenom"`                                                                // Prénom
	Email             string    `json:"email" gorm:"unique"`                                                   // Email (unique)
	Genre             string    `json:"genre"`                                                                 // Sexe ou genre
	Telephone         string    `json:"telephone"`                                                             // Numéro de téléphone
	MotDePasseHash    string    `json:"-"`                                                                     // Mot de passe hashé (non exporté en JSON)
	Role              string    `json:"role"`                                                                  // Rôle de l'utilisateur (admin, opérateur, etc.)
	Permissions       []string  `json:"permissions"`                                                           // Liste des permissions
	TenantID          uuid.UUID `json:"tenant_id"`                                                             // Identifiant du client (multi-tenant)
	Status            string    `json:"status"`                                                                // Statut du compte (active, suspended, locked, etc.)
	IsActive          bool      `json:"is_active"`                                                             // Compte actif ou non
	DateInscription   time.Time `json:"date_inscription"`                                                      // Date d'inscription
	DerniereConnexion time.Time `json:"derniere_connexion"`                                                    // Dernière date de connexion
	TokenExp          time.Time `json:"token_exp"`                                                             // Expiration du token JWT
	MFAEnabled        bool      `json:"mfa_enabled"`                                                           // Authentification multi-facteur activée
	PhotoProfil       string    `json:"photo_profil"`                                                          // URL de la photo de profil
}

type UtilisateurRedis struct {
	IDUtilisateur string
	Nom           string
	Prenom        string
	Email         string
	Telephone     string
	Role          string
	Permissions   string
	TenantID      string
	MFAEnabled    bool
	IsActive      bool
	Status        string
}
