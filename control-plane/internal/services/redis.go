package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/logger"
	"go.uber.org/zap"
)

var RedisClient *redis.Client

const nodeCacheTTL = 15 * time.Minute

func InitRedis(cfg *config.Config) {

	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf(" Impossible de se connecter à Redis: %v", err)
	} else {
		log.Printf(" Connexion Redis réussie: %s", pong)
	}
}

func CacheNode(ctx context.Context, node *repository.Node) error {
	key := fmt.Sprintf("node:%s", node.ID)

	data, err := json.Marshal(node)
	if err != nil {
		logger.Log.Error("Erreur JSON lors du cache node", zap.String("node_id", node.ID), zap.Error(err))
		return err
	}

	err = RedisClient.Set(ctx, key, data, nodeCacheTTL).Err()
	if err != nil {
		logger.Log.Error("Échec cache node Redis", zap.String("key", key), zap.Error(err))
		return err
	}

	logger.Log.Debug("Node mis en cache", zap.String("node_id", node.ID))
	return nil
}

func GetUserInfoFromRedis(ctx context.Context, userID string) (*repository.UtilisateurRedis, error) {
	redisKey := fmt.Sprintf("user:session:%s", userID)

	//  toutes les valeurs stockées sous cette clé
	data, err := RedisClient.HGetAll(ctx, redisKey).Result()
	if err != nil {
		log.Printf("Erreur Redis lors de la récupération des infos utilisateur [ID: %s]: %v", userID, err)
		return nil, err
	}

	if len(data) == 0 {
		log.Printf(" Aucun enregistrement trouvé en Redis pour [ID: %s]", userID)
		return nil, fmt.Errorf("utilisateur non trouvé en cache")
	}

	// Convertir les valeurs récupérées en une struct `Utilisateur`
	user := &repository.UtilisateurRedis{
		IDUtilisateur: data["id_utilisateur"],
		Nom:           data["nom"],
		Prenom:        data["prenom"],
		Email:         data["email"],
		Telephone:     data["telephone"],
		Role:          data["role"],
		Permissions:   data["permissions"],
		TenantID:      data["id_tenant"],
		MFAEnabled:    data["mfa_enabled"] == "true",
	}

	log.Printf("Infos utilisateur récupérées depuis Redis [ID: %s]", userID)
	return user, nil
}
