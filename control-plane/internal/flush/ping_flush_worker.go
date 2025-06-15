// ping_flush_worker implémente un worker de fond qui lit les informations de "last_seen" et "status"
// des nœuds enregistrés dans Redis, puis les insère ou met à jour en base PostgreSQL.
// Il est déclenché automatiquement toutes les 2 minutes via un planificateur cron.
//
// Objectifs :
//
//	Réduire la pression d’écriture directe sur PostgreSQL provoquée par chaque Ping().
//	Éviter les écritures redondantes si les données n’ont pas changé.
//	Garantir une persistance fiable des dernières activités des worker-nodes.
package workers

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
)

const (
	lastSeenPrefix = "node:last_seen:"
	statusPrefix   = "node:status:"
)

// StartPingFlushWorker lance un job cron toutes les 2 minutes.
// Ce job lit les valeurs last_seen et status de Redis et les insère en base PostgreSQL.
// Cette fonction est non-bloquante et doit être appelée au démarrage du control-plane.
func StartPingFlushWorker(repo repository.NodeRepository) {
	c := cron.New()
	c.AddFunc("@every 2m", func() {
		FlushHeartbeatToPostgres(repo)
	})
	c.Start()
	log.Println("[PING FLUSH WORKER] Cron started (interval: 2m)")
}

// FlushHeartbeatToPostgres récupère les pings récents stockés dans Redis
// et les met à jour en base de données PostgreSQL pour chaque node concerné.
// Il met également à jour le statut si une valeur est présente.
//
// Clés attendues dans Redis :
//
//	node:last_seen:{nodeID} => timestamp RFC3339
//	node:status:{nodeID}    => string ("online", "offline", "degraded")
func FlushHeartbeatToPostgres(repo repository.NodeRepository) {
	ctx := context.Background()
	keys, err := pkg.RedisClient.Keys(ctx, lastSeenPrefix+"*").Result()
	if err != nil {
		log.Printf("[PING FLUSH] Erreur récupération des clés Redis: %v\n", err)
		return
	}

	for _, key := range keys {
		nodeID := strings.TrimPrefix(key, lastSeenPrefix)
		lastSeenStr, err := pkg.RedisClient.Get(ctx, key).Result()
		if err != nil {
			log.Printf("[PING FLUSH] Erreur récupération valeur pour %s: %v\n", key, err)
			continue
		}

		parsedTime, err := time.Parse(time.RFC3339, lastSeenStr)
		if err != nil {
			log.Printf("[PING FLUSH] Erreur parsing time %s: %v\n", lastSeenStr, err)
			continue
		}

		// met à jour en DB
		err = repo.UpdateHeartbeat(ctx, nodeID, parsedTime)
		if err != nil {
			log.Printf("[PING FLUSH] Erreur DB heartbeat pour %s: %v\n", nodeID, err)
		} else {
			log.Printf("[PING FLUSH] Heartbeat maj pour node %s à %s\n", nodeID, parsedTime.Format(time.RFC822))
		}

		// Récupération du statut si disponible
		statusKey := statusPrefix + nodeID
		if status, err := pkg.RedisClient.Get(ctx, statusKey).Result(); err == nil {
			err := repo.SetNodeStatus(ctx, nodeID, status)
			if err != nil {
				log.Printf("[PING FLUSH] Erreur mise à jour status %s: %v\n", nodeID, err)
			} else {
				log.Printf("[PING FLUSH]  Status maj pour node %s à %s\n", nodeID, status)
			}
		}
	}
}
