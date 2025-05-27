package services

import (
	"fmt"
	"log"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	"gopkg.in/gomail.v2"
)

// envoie une alerte de sécurité par email
func SendSecurityAlerteEmail(email, ipAddress, userAgent string) error {
	log.Printf("Alerte de sécurité détectée pour %s", email)

	subject := "Alerte de connexion suspecte détectée"
	body := fmt.Sprintf(`
	Bonjour,

	Nous avons détecté une tentative de connexion à votre compte depuis un nouvel appareil.

	🔹 **Adresse IP:** %s  
	🔹 **Appareil:** %s  
	🔹 **Heure:** %s  

	Si cette connexion est légitime, vous pouvez ignorer cet email.  
	Si ce n'est pas vous, nous vous recommandons de **changer votre mot de passe immédiatement**.

	Sécurité avant tout !
	`, ipAddress, userAgent, time.Now().Format("2006-01-02 15:04:05"))

	err := sendEmail(email, subject, body)
	if err != nil {
		log.Printf("Échec de l'envoi de l'email d'alerte à %s: %v", email, err)
		return fmt.Errorf("échec de l'envoi de l'email d'alerte")
	}

	log.Printf("Alerte de sécurité envoyée par email à %s", email)
	return nil
}

func sendEmail(to, subject, body string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Erreur de chargement de la configuration pour l'envoi d'email : %v", err)
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.Email.SMTPUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.SMTPUser, cfg.Email.SMTPPassword)
	d.SSL = true

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Erreur d'envoi de l'email à %s: %v", to, err)
		return err
	}

	log.Printf("Email envoyé avec succès à %s", to)
	return nil
}
