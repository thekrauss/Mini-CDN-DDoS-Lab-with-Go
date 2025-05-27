package services

import (
	"fmt"
	"log"
	"time"
)

func (s *AuthService) SendAdminSecurityAlertEmail(adminEmail, email, nom, prenom, identifiant, password, organisme, ip, userAgent string) {
	go func() {
		err := SendAdminCreatedAlert([]string{adminEmail}, email, nom, prenom, identifiant, password, organisme, ip, userAgent)
		if err != nil {
			log.Printf("Erreur lors de l'envoi du mail de sécurité  Admin : %v", err)
		}
	}()
}

func SendAdminCreatedAlert(adminEmails []string, email, nom, prenom, identifiant, password, organisme, ip, userAgent string) error {
	metadata := time.Now().Format("2006-01-02 15:04:05")

	// Email aux administrateurs
	subject := fmt.Sprintf("Nouvel administrateur  : %s %s", nom, prenom)
	body := fmt.Sprintf(`
	Bonjour,

	Un nouvel administrateur a été enregistré dans le système .

	👤 Nom complet : %s %s
	🏫 Organisation : %s
	📧 Email : %s
	🌍 IP : %s
	🧭 User-Agent : %s
	📅 Date : %s

	Merci de vérifier ces informations dans votre tableau de bord.

	Cordialement,
	L'équipe`, nom, prenom, organisme, email, ip, userAgent, metadata)

	for _, admin := range adminEmails {
		if err := sendEmail(admin, subject, body); err != nil {
			log.Printf("Erreur lors de l'envoi à l'admin %s : %v", admin, err)
		}
	}

	// Email à l'utilisateur
	return sendUserWelcomeEmail(email, nom, prenom, identifiant, password)
}

// sendUserWelcomeEmail envoie un email de bienvenue avec les identifiants
func sendUserWelcomeEmail(email, nom, prenom, identifiant, password string) error {
	subject := "🎓 Bienvenue sur SYK - Votre compte est prêt"
	body := fmt.Sprintf(`
	Bonjour %s %s,

	Votre compte administrateur CDN été créé avec succès.

	🔐 Identifiants : %s
	📧 Email : %s
	🔑 Mot de passe : %s

	💡 Veuillez changer votre mot de passe dès votre première connexion.
	🔗 Connexion : https://cdn.com/login

	Cordialement,
	L'équipe`, nom, prenom, identifiant, email, password)

	return sendEmail(email, subject, body)
}
