package jwt

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
	"google.golang.org/grpc/metadata"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
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

func VerifyJWT(tokenString string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.JWT.RefreshSecret), nil
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
