package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Type de clé pour éviter les conflits dans le contexte
type contextKey string

const userClaimsKey contextKey = "userClaims"

// vérifie le JWT et stocke les claims directement dans le contexte
func AuthMiddleware(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		// Autorise certains endpoints sans authentification (exemple Login)
		if strings.Contains(info.FullMethod, "Login") {
			return handler(ctx, req)
		}

		// Récupère les métadonnées gRPC
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("métadonnées manquantes")
		}

		authHeader, exists := md["authorization"]
		if !exists || len(authHeader) == 0 {
			return nil, fmt.Errorf("token d'authentification manquant")
		}

		// Extraire le token (Bearer)
		tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")

		// Valider et parser le token
		claims, err := auth.ValidateJWT(tokenString, cfg.JWT.SecretKey)
		if err != nil {
			return nil, fmt.Errorf("token invalide : %v", err)
		}

		// Injecter les claims directement dans le contexte
		ctx = context.WithValue(ctx, userClaimsKey, claims)

		// Poursuivre la chaîne d'interception
		return handler(ctx, req)
	}
}

// TimeoutMiddleware limite la durée d'une requête
func TimeoutMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		return handler(ctx, req)
	}
}
