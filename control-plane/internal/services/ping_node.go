package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/monitoring"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ping reçoit les battements de cœur (heartbeats) envoyés périodiquement par un agent worker-node.
//
// Cette méthode est appelée automatiquement toutes les 10 secondes par chaque nœud actif.
// Elle est essentielle au suivi en temps réel de la santé de l'infrastructure distribuée.
//
// Elle effectue notamment :
//  La mise à jour de la clé "last_seen" du nœud dans Redis (cache volatile).
//  Le calcul dynamique du statut du nœud "online", "degraded" en fonction des métriques système (CPU, mémoire).
//
//  Optimisation :
// Pour éviter une charge excessive sur PostgreSQL, les informations sont d’abord stockées dans Redis.
// Une écriture en base de données n’est déclenchée que si :
//  Le statut du nœud a changé (ex. : passage de "degraded" à "online").
//  Le dernier ping reçu remonte à plus de 30 secondes (rafraîchissement nécessaire).
//
//  Expiration automatique :
// Les clés Redis sont configurées avec un TTL (Time-To-Live) de 2 minutes.
// Si un nœud ne ping plus après ce délai, il est considéré comme "offline".
//
//  Persistance :
// Un processus asynchrone (scheduler / cron worker) est chargé de flusher les données de Redis vers PostgreSQL à intervalles réguliers.

func (s *NodeService) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {

	// authentification par contexte
	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}

	if err := s.CheckAdminPermissions(ctx, claims, PermPingNode); err != nil {
		return nil, err
	}

	monitoring.PingCounter.Inc()

	if req.NodeId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_id requis")
	}

	node, err := s.Repo.GetNodeByID(ctx, req.NodeId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Node non trouvé: %v", err)
	}

	now := time.Now()
	cfg := s.Config.MonitoringEtat

	cpu := float64(req.Cpu)
	mem := float64(req.Memory)

	// Calcul du statut
	var newStatus repository.NodeStatus
	switch {
	case cpu >= cfg.CriticalCPUThreshold || mem >= cfg.CriticalMemThreshold:
		newStatus = repository.NodeDegraded
	case cpu >= cfg.DegradedCPUThreshold || mem >= cfg.DegradedMemThreshold:
		newStatus = repository.NodeDegraded
	default:
		newStatus = repository.NodeOnline
	}

	// Anti-spam ping: ignore si trop fréquent (5s)
	lastPingKey := fmt.Sprintf("node:last_ping:%s", node.ID)
	if lastPingStr, err := pkg.RedisClient.Get(ctx, lastPingKey).Result(); err == nil {
		if parsed, err := time.Parse(time.RFC3339, lastPingStr); err == nil {
			if time.Since(parsed) < 5*time.Second {
				return &pb.PingResponse{Status: "skipped"}, nil
			}
		}
	}
	_ = pkg.RedisClient.Set(ctx, lastPingKey, now.Format(time.RFC3339), 30*time.Second)

	// TTL Redis ajusté
	ttl := 2 * time.Minute
	if newStatus == repository.NodeDegraded {
		ttl = 5 * time.Minute
	}

	keyLastSeen := fmt.Sprintf("node:last_seen:%s", node.ID)
	keyStatus := fmt.Sprintf("node:status:%s", node.ID)

	// Cache enrichi (info CPU/MEM)
	_ = pkg.RedisClient.Set(ctx, keyLastSeen, now.Format(time.RFC3339), ttl)
	_ = pkg.RedisClient.Set(ctx, keyStatus, string(newStatus), ttl)
	_ = pkg.RedisClient.HSet(ctx, "node:info:"+node.ID, map[string]any{
		"last_seen": now.Format(time.RFC3339),
		"status":    newStatus,
		"cpu":       cpu,
		"memory":    mem,
		"uptime":    req.UptimeSeconds,
	})

	// met à jour en base uniquement si nécessaire
	shouldUpdateStatus := node.Status != string(newStatus)
	shouldUpdateHeartbeat := now.Sub(node.LastSeen) > 30*time.Second

	if shouldUpdateStatus {
		if err := s.Repo.SetNodeStatus(ctx, node.ID, string(newStatus)); err != nil {
			return nil, status.Errorf(codes.Internal, "Erreur mise à jour statut: %v", err)
		}
	}

	if shouldUpdateHeartbeat {
		if err := s.Repo.UpdateHeartbeat(ctx, node.ID, now); err != nil {
			return nil, status.Errorf(codes.Internal, "Erreur mise à jour heartbeat: %v", err)
		}
	}

	// enregistrement des métriques (async)
	metrics := &repository.NodeMetrics{
		NodeID:    node.ID,
		TenantID:  node.TenantID,
		Timestamp: now,
		CPU:       cpu,
		Memory:    mem,
		Uptime:    int64(req.UptimeSeconds),
		Status:    string(newStatus),
	}
	if raw, err := json.Marshal(metrics); err == nil {
		_ = pkg.RedisClient.Set(ctx, "node:metrics:"+node.ID, raw, 5*time.Minute)
	}

	if s.Hub != nil {
		update := map[string]any{
			"node_id": node.ID,
			"status":  newStatus,
			"cpu":     cpu,
			"mem":     mem,
			"uptime":  req.UptimeSeconds,
			"seen_at": now.Format(time.RFC3339),
		}
		if raw, err := json.Marshal(update); err == nil {
			s.Hub.Broadcast(raw)
		}
	}

	// notification si dégradé (Pub/Sub, webhook, Discord)
	if newStatus == repository.NodeDegraded {
		log.Printf("Node %s dégradé - CPU: %.2f%% MEM: %.2f%%", node.ID, cpu, mem)

	}

	return &pb.PingResponse{
		Status: "ok",
	}, nil
}
