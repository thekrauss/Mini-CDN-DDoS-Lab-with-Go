package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	"google.golang.org/grpc/metadata"
)

// Custom claims can include user-specific data such as ID, role, and expiration time.
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token with user-specific claims
func GenerateJWT(userID uuid.UUID, username, role string) (string, error) {
	secretKey, err := config.GetSecret("JWT_REFRESH_SECRET")
	if err != nil {
		return "", fmt.Errorf("failed to get JWT_REFRESH_SECRET: %w", err)
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func VerifyJWT(tokenString string) (*Claims, error) {
	secretKey, err := config.GetSecret("JWT_REFRESH_SECRET")
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT_REFRESH_SECRET: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}

	return claims, nil
}

func ValidateJWT(tokenString, secretKey string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signature method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token invalide ou expiré")
}

func ExtractJWTFromContext(ctx context.Context) (*Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("aucun token trouvé dans les métadonnées")
	}

	//  le token depuis le header "authorization"
	authHeader, exists := md["authorization"]
	if !exists || len(authHeader) == 0 {
		return nil, fmt.Errorf("authorization header manquant")
	}

	tokenString := authHeader[0]
	if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
		return nil, fmt.Errorf("format du token invalide")
	}

	// Extraire la valeur réelle du token
	tokenString = tokenString[7:]

	// verifie et extraire les claims
	claims, err := VerifyJWT(tokenString)
	if err != nil {
		return nil, fmt.Errorf("échec de vérification du JWT: %w", err)
	}

	return claims, nil
}
