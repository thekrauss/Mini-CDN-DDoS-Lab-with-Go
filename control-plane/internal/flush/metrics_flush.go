package workers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
)

func StartMetricsFlushWorker(repo repository.NodeRepository) {
	ticker := time.NewTicker(2 * time.Minute) // flush toutes les 2 min

	go func() {
		for range ticker.C {
			ctx := context.Background()

			// get toutes les clés Redis contenant des métriques
			keys, err := pkg.RedisClient.Keys(ctx, "node:metrics:*").Result()
			if err != nil {
				log.Printf("[FLUSH METRICS] Erreur récupération des clés Redis: %v", err)
				continue
			}

			for _, key := range keys {
				raw, err := pkg.RedisClient.Get(ctx, key).Result()
				if err != nil {
					log.Printf("[FLUSH METRICS] Erreur lecture %s: %v", key, err)
					continue
				}

				var metrics repository.NodeMetrics
				if err := json.Unmarshal([]byte(raw), &metrics); err != nil {
					log.Printf("[FLUSH METRICS] JSON invalide: %v", err)
					continue
				}

				if err := repo.StoreNodeMetrics(ctx, &metrics); err != nil {
					log.Printf("[FLUSH METRICS] Insertion PostgreSQL échouée: %v", err)
					continue
				}

				// delete la clé si insertion réussie
				_ = pkg.RedisClient.Del(ctx, key).Err()
			}

			log.Printf("[FLUSH METRICS] Flush terminé (%d métriques)", len(keys))
		}
	}()
}
