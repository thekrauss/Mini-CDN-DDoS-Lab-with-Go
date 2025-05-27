package pkg

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
)

var RedisClient *redis.Client

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
