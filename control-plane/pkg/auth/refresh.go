package auth

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v4"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
)

// génère un refresh token
func GenerateRefreshToken(userID uuid.UUID) (string, error) {
	secretKey, err := config.GetSecret("JWT_REFRESH_SECRET")
	if err != nil {
		return "", fmt.Errorf("failed to get JWT_REFRESH_SECRET: %w", err)
	}

	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 jours
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "auth-service",
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// vérifie si un refresh token est valide
func VerifyRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	secretKey, err := config.GetSecret("JWT_REFRESH_SECRET")
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT_REFRESH_SECRET: %w", err)
	}
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	return claims, nil
}
