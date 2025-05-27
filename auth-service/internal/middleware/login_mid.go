package middleware

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/services"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var rateLimit = make(map[string]int)
var mutex = sync.Mutex{}

// requêtes par minute par utilisateur
const requestLimit = 10
const duration = time.Minute

func LoggingMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		// Journalise la méthode appelée
		log.Printf("Requête entrante : %s", info.FullMethod)

		// execute le handler
		resp, err := handler(ctx, req)

		// Journalise la durée d'exécution
		duration := time.Since(start)
		log.Printf("Requête traitée : %s | Durée : %s", info.FullMethod, duration)

		return resp, err
	}
}

func RateLimitingMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		ip, _ := services.GetRequestMetadata(ctx)

		// le nombre de requêtes
		mutex.Lock()
		defer mutex.Unlock()

		if rateLimit[ip] >= requestLimit {
			return nil, errors.New("trop de requêtes, veuillez réessayer plus tard")
		}

		// Augmenter le compteur
		rateLimit[ip]++
		go resetRateLimit(ip)

		return handler(ctx, req)
	}
}

func CheckPermissionMiddleware(service *services.AuthService, requiredPermission string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		claims, err := auth.ExtractJWTFromContext(ctx)
		if err != nil {
			log.Printf("Erreur JWT: %v", err)
			return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
		}

		// vérification des permissions
		hasPerm, err := service.HasSykPermission(ctx, claims.UserID.String(), requiredPermission)
		if err != nil {
			log.Printf("Erreur lors de la vérification des permissions: %v", err)
			return nil, status.Errorf(codes.Internal, "Erreur interne")
		}
		if !hasPerm {
			log.Printf("Permission refusée: %s", requiredPermission)
			return nil, status.Errorf(codes.PermissionDenied, "Vous n'avez pas la permission d'effectuer cette action")
		}

		// l'utilisateur a la permission, on continue l'exécution de la requête
		return handler(ctx, req)
	}
}

// Réinitialise le compteur après une durée donnée
func resetRateLimit(ip string) {
	time.Sleep(duration)
	mutex.Lock()
	defer mutex.Unlock()
	rateLimit[ip] = 0
}
