package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
)

// Structure de la réponse Firebase
type resetResponse struct {
	OobLink string `json:"oobLink"`
}

// Génère un lien de réinitialisation de mot de passe via Firebase API REST
func GenerateResetLinkREST(ctx context.Context, email string) (string, error) {

	apiKey := config.AppConfig.Firebase.FirebaseAPIKey
	if apiKey == "" {
		return "", fmt.Errorf("Firebase API Key manquante")
	}

	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:sendOobCode?key=%s", apiKey)

	reqBody, err := json.Marshal(map[string]interface{}{
		"requestType": "PASSWORD_RESET",
		"email":       email,
	})
	if err != nil {
		return "", fmt.Errorf("Erreur lors de la construction de la requête : %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("Erreur lors de la création de la requête HTTP : %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Erreur lors de l'appel API Firebase : %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Erreur lors de la lecture de la réponse : %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Firebase API Error: %s", body)
	}

	var res resetResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("Erreur lors du parsing de la réponse JSON : %w", err)
	}

	return res.OobLink, nil
}
