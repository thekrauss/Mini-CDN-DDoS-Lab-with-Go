package services

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App

func InitFirebase() (*firebase.App, error) {
	credFile, err := SaveFirebaseCredentialsToFile()
	if err != nil {
		return nil, fmt.Errorf("Impossible d'écrire les credentials Firebase : %v", err)
	}
	defer os.Remove(credFile)

	opt := option.WithCredentialsFile(credFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("Erreur d'initialisation Firebase : %v", err)
	}

	FirebaseApp = app
	return app, nil
}

func SaveFirebaseCredentialsToFile() (string, error) {
	credentials := config.AppConfig.Firebase.FirebaseCredentials

	if credentials == "" {
		return "", fmt.Errorf("Firebase credentials non définis")
	}

	tmpFile, err := os.CreateTemp("", "firebase-*.json")
	if err != nil {
		return "", fmt.Errorf("Erreur lors de la création du fichier temporaire : %v", err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write([]byte(credentials))
	if err != nil {
		return "", fmt.Errorf("Erreur lors de l'écriture des credentials Firebase : %v", err)
	}

	return tmpFile.Name(), nil
}
