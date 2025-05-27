package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/db"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/services"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	Store   *db.DBStore
	Service *services.AuthService
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

	log.Printf("Received login request: identifier=%s", services.MaskSensitiveData(req.Identifier))

	identifier := strings.ToLower(strings.TrimSpace(req.Identifier))
	password := strings.TrimSpace(req.Password)
	ipAddress := req.IpAddress
	userAgent := req.UserAgent

	if identifier == "" || password == "" {
		return nil, fmt.Errorf("identifier and password are required")
	}

	maskedIdentifier := services.MaskSensitiveData(identifier)
	log.Printf("Received login request for identifier=%s, IP=%s", maskedIdentifier, ipAddress)

	ctxRedis, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	redisKey := fmt.Sprintf("login_attempts:%s", identifier)

	// si l'utilisateur est bloqué
	blocked, err := services.RedisClient.Get(ctxRedis, fmt.Sprintf("blocked:%s", identifier)).Result()
	if err == nil && blocked == "true" {
		log.Printf(" User %s is blocked due to too many failed login attempts.", maskedIdentifier)
		return nil, fmt.Errorf("too many failed login attempts Try again later")
	}

	attempts, err := services.RedisClient.Get(ctxRedis, redisKey).Int()
	if err != nil && err != redis.Nil {
		log.Printf("Erreur d'accès à Redis pour la clé %s: %v", redisKey, err)
	}

	log.Printf("Nombre de tentatives pour %s: %d", maskedIdentifier, attempts)

	if attempts >= config.AppConfig.Security.MaxFailedAttempts {
		log.Printf("User %s is temporarily blocked.", maskedIdentifier)
		services.RedisClient.Set(ctxRedis, fmt.Sprintf("blocked:%s", identifier), "true", config.AppConfig.Security.LockoutDuration)
		services.RedisClient.Del(ctxRedis, redisKey)
		return nil, fmt.Errorf("too many failed login attempts. Try again later")
	}

	token, userID, role, err := services.AuthenticateUser(s.Store.DB, identifier, password, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, services.ErrMFARequired) {
			return &pb.LoginResponse{
				Message:     "MFA required, please verify OTP",
				RequiresOtp: true,
			}, nil
		}

		// incrémente le compteur d'échecs en Redis
		services.RedisClient.Incr(ctxRedis, redisKey)
		services.RedisClient.Expire(ctxRedis, redisKey, 15*time.Minute)

		// alerte après 3 tentatives échouées
		if attempts >= 3 {
			go services.SendSecurityAlerteEmail(identifier, ipAddress, userAgent)
		}

		log.Printf(" Authentication failed for %s. Attempts: %d", maskedIdentifier, attempts)
		return nil, fmt.Errorf("incorrect identifier or password")
	}

	services.RedisClient.Del(ctxRedis, redisKey)

	refreshToken, err := auth.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	err = s.Service.SaveRefreshToken(userID, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	user, err := s.Service.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Impossible de récupérer les infos ... pour %s: %v", userID.String(), err)
	} else {
		go func() {
			err := services.StoreUserSessionInRedis(ctxRedis, user, ipAddress, userAgent)
			if err != nil {
				log.Printf("Impossible de stocker la session utilisateur dans Redis: %v", err)
			}
		}()
	}

	validatedRole, err := services.ParseRole(role)
	if err != nil {
		return nil, fmt.Errorf("accès interdit : rôle %s non autorisé", role)
	}

	log.Printf("Utilisateur connecté avec un rôle valide : %s", validatedRole)

	//  met à jour la dernière activité
	services.UpdateLastActivity(s.Store.DB, userID)

	log.Printf("User %s logged in successfully, userID: %s, role: %s", maskedIdentifier, userID.String(), role)

	return &pb.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Message:      "Login successful",
		Role:         validatedRole,
		UserId:       userID.String(),
	}, nil
}
