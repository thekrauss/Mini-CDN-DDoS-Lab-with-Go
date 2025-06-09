// internal/db/node_sql.go
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/logger"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
	"go.uber.org/zap"
)

type SQLNodeRepository struct {
	DB *sql.DB
}

func NewNodeRepository(db *sql.DB) repository.NodeRepository {
	return &SQLNodeRepository{DB: db}
}

// --- CRUD ---

func (r *SQLNodeRepository) CreateNode(ctx context.Context, node *repository.Node) error {
	query := `
  INSERT INTO nodes (id, hostname, ip_address, tenant_id, last_seen, created_at, updated_at, location, os, version, status, provider, is_blacklisted, tags)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
`
	_, err := r.DB.ExecContext(ctx, query,
		node.ID, node.Name, node.IP, node.TenantID, node.LastSeen,
		node.CreatedAt, node.UpdatedAt,
		node.Location, node.OS, node.SoftwareVersion,
		node.Status, node.Provider, node.IsBlacklisted, pq.Array(node.Tags),
	)

	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("nodes:tenant:%s", node.TenantID)
	if delErr := pkg.RedisClient.Del(ctx, cacheKey).Err(); delErr != nil {
		logger.Log.Warn("failed to invalidate tenant node list cache", zap.String("tenant_id", node.TenantID), zap.Error(delErr))
	}

	return nil
}

func (r *SQLNodeRepository) GetNodeByID(ctx context.Context, id string) (*repository.Node, error) {
	cachedKey := fmt.Sprintf("node:%s", id)

	cached, err := pkg.RedisClient.Get(ctx, cachedKey).Result()
	if err == nil {
		var node repository.Node
		if err := json.Unmarshal([]byte(cached), &node); err == nil {
			return &node, nil
		}
		logger.Log.Warn("cache corrupted", zap.String("node_id", id), zap.Error(err))
	}

	query := `SELECT id, hostname, ip_address, tenant_id, status, last_seen, created_at, updated_at, location, os, version, provider, is_blacklisted, tags FROM nodes WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, id)

	var node repository.Node
	var tags []string
	if err := row.Scan(&node.ID, &node.Name, &node.IP, &node.TenantID, &node.Status, &node.LastSeen, &node.CreatedAt, &node.UpdatedAt, &node.Location, &node.OS, &node.SoftwareVersion, &node.Provider, &node.IsBlacklisted, pq.Array(&tags)); err != nil {
		return nil, err
	}
	node.Tags = tags

	if raw, err := json.Marshal(&node); err == nil {
		if err := pkg.RedisClient.Set(ctx, cachedKey, raw, 60*time.Minute).Err(); err != nil {
			logger.Log.Warn("Failed to set Redis cache", zap.String("key", cachedKey), zap.Error(err))
		}
	}
	return &node, nil
}

func (r *SQLNodeRepository) UpdateHeartbeat(ctx context.Context, id string, seenAt time.Time) error {
	query := `UPDATE nodes SET last_seen = $1, updated_at = $1 WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, seenAt, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("node:%s", id)
	if delErr := pkg.RedisClient.Del(ctx, cacheKey).Err(); delErr != nil {
		logger.Log.Warn("failed to invalidate node cache", zap.String("node_id", id), zap.Error(delErr))
	}

	return nil
}

func (r *SQLNodeRepository) ListNodesByTenant(ctx context.Context, tenantID string) ([]*repository.Node, error) {
	cachedKey := fmt.Sprintf("nodes:tenant:%s", tenantID)

	cached, err := pkg.RedisClient.Get(ctx, cachedKey).Result()
	if err == nil {
		var nodes []*repository.Node
		if err := json.Unmarshal([]byte(cached), &nodes); err == nil {
			return nodes, nil
		}
		logger.Log.Warn("cache corrompu pour liste de nodes",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
	}

	query := `SELECT id, hostname, ip_address, tenant_id, status, last_seen, created_at, updated_at, location, os, version, provider, is_blacklisted, tags FROM nodes WHERE tenant_id = $1`
	rows, err := r.DB.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*repository.Node
	for rows.Next() {
		var node repository.Node
		var tags []string
		if err := rows.Scan(&node.ID, &node.Name, &node.IP, &node.TenantID, &node.Status, &node.LastSeen, &node.CreatedAt, &node.UpdatedAt, &node.Location, &node.OS, &node.SoftwareVersion, &node.Provider, &node.IsBlacklisted, pq.Array(&tags)); err != nil {
			return nil, err
		}
		node.Tags = tags
		nodes = append(nodes, &node)
	}

	if raw, err := json.Marshal(nodes); err == nil {
		_ = pkg.RedisClient.Set(ctx, cachedKey, raw, 15*time.Minute).Err()
	}

	return nodes, nil
}

func (r *SQLNodeRepository) AssignToTenant(ctx context.Context, nodeID, tenantID string) error {
	query := `UPDATE nodes SET tenant_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, tenantID, nodeID)
	return err
}

func (r *SQLNodeRepository) DeleteNode(ctx context.Context, id string) error {
	var tenantID string
	err := r.DB.QueryRowContext(ctx, "SELECT tenant_id FROM nodes WHERE id = $1", id).Scan(&tenantID)
	if err != nil {
		return err
	}

	_, err = r.DB.ExecContext(ctx, "DELETE FROM nodes WHERE id = $1", id)
	if err != nil {
		return err
	}

	_ = pkg.RedisClient.Del(ctx, fmt.Sprintf("node:%s", id)).Err()
	_ = pkg.RedisClient.Del(ctx, fmt.Sprintf("nodes:tenant:%s", tenantID)).Err()

	return nil
}

func (r *SQLNodeRepository) UpdateNodeMetadata(ctx context.Context, id string, name string, ip string, tags map[string]string) error {

	query := `UPDATE nodes SET name = $1, ip = $2, tags = $3, updated_at = NOW() WHERE id = $4`
	tagJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, query, name, ip, tagJSON, id)
	return err
}

func (r *SQLNodeRepository) SearchNodes(ctx context.Context, filter repository.NodeFilter) ([]*repository.Node, error) {
	cachedKey := fmt.Sprintf("search:tenant:%s:q:%s", filter.TenantID, filter.Query)

	cached, err := pkg.RedisClient.Get(ctx, cachedKey).Result()
	if err == nil {
		var nodes []*repository.Node
		if err := json.Unmarshal([]byte(cached), &nodes); err == nil {
			return nodes, nil
		}
		logger.Log.Warn("SearchNodes cache corrompu", zap.String("key", cachedKey), zap.Error(err))
	}

	query := `
		SELECT id, name, ip, tenant_id, status, last_seen, created_at, updated_at
		FROM nodes
		WHERE tenant_id = $1 AND (name ILIKE $2 OR ip ILIKE $2)
	`
	rows, err := r.DB.QueryContext(ctx, query, filter.TenantID, "%"+filter.Query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*repository.Node
	for rows.Next() {
		var node repository.Node
		if err := rows.Scan(&node.ID, &node.Name, &node.IP, &node.TenantID, &node.Status, &node.LastSeen, &node.CreatedAt, &node.UpdatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	if raw, err := json.Marshal(nodes); err == nil {
		_ = pkg.RedisClient.Set(ctx, cachedKey, raw, 5*time.Minute).Err()
	}

	return nodes, nil
}

func (r *SQLNodeRepository) CountActiveNodes(ctx context.Context, tenantID string, since time.Duration) (int, error) {
	cachedKey := fmt.Sprintf("count:active:tenant:%s:since:%s", tenantID, since.String())

	cached, err := pkg.RedisClient.Get(ctx, cachedKey).Result()
	if err == nil {
		var count int
		if err := json.Unmarshal([]byte(cached), &count); err == nil {
			return count, nil
		}
	}

	query := `SELECT COUNT(*) FROM nodes WHERE tenant_id = $1 AND last_seen > NOW() - $2::interval`
	var count int
	err = r.DB.QueryRowContext(ctx, query, tenantID, since.String()).Scan(&count)
	if err != nil {
		return 0, err
	}

	if raw, err := json.Marshal(count); err == nil {
		_ = pkg.RedisClient.Set(ctx, cachedKey, raw, 2*time.Minute).Err()
	}

	return count, nil
}

// --- Statut / Orchestration ---

func (r *SQLNodeRepository) SetNodeStatus(ctx context.Context, id string, status string) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE nodes SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}

func (r *SQLNodeRepository) GetInactiveNodes(ctx context.Context, olderThan time.Duration) ([]*repository.Node, error) {
	cachedKey := fmt.Sprintf("inactive:nodes:since:%s", olderThan.String())

	if cached, err := pkg.RedisClient.Get(ctx, cachedKey).Result(); err == nil {
		var nodes []*repository.Node
		if err := json.Unmarshal([]byte(cached), &nodes); err == nil {
			return nodes, nil
		}
		logger.Log.Warn("GetInactiveNodes cache corrompu", zap.Error(err))
	}

	query := `
		SELECT id, name, ip, tenant_id, status, last_seen, created_at, updated_at
		FROM nodes
		WHERE last_seen < NOW() - $1::interval
	`
	rows, err := r.DB.QueryContext(ctx, query, olderThan.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*repository.Node
	for rows.Next() {
		var node repository.Node
		if err := rows.Scan(&node.ID, &node.Name, &node.IP, &node.TenantID, &node.Status, &node.LastSeen, &node.CreatedAt, &node.UpdatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	if raw, err := json.Marshal(nodes); err == nil {
		_ = pkg.RedisClient.Set(ctx, cachedKey, raw, 2*time.Minute).Err()
	}

	return nodes, nil
}

func (r *SQLNodeRepository) MarkAllNodesOffline(ctx context.Context) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE nodes SET status = 'offline', updated_at = NOW()")
	return err
}

// --- Sécurité / IP ---

func (r *SQLNodeRepository) IsIPAlreadyRegistered(ctx context.Context, ip string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM nodes WHERE ip = $1)`
	var exists bool
	err := r.DB.QueryRowContext(ctx, query, ip).Scan(&exists)
	return exists, err
}

func (r *SQLNodeRepository) InvalidateTenantCache(ctx context.Context, tenantID string) error {
	cacheKey := fmt.Sprintf("nodes:tenant:%s", tenantID)
	err := pkg.RedisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		logger.Log.Warn("invalidate cache tenant",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
	}
	return err
}

func (r *SQLNodeRepository) InsertAuditLog(ctx context.Context, log *repository.AuditLog) error {
	query := `
		INSERT INTO infra_audit_logs 
			(id, user_id, role, action, target, details, ip_address, user_agent, timestamp, tenant_id) 
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.DB.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.Role,
		log.Action,
		log.Target,
		log.Details,
		log.IPAddress,
		log.UserAgent,
		log.Timestamp,
		log.TenantID,
	)

	return err
}

func (r *SQLNodeRepository) GetAuditLogs(ctx context.Context, filter repository.AuditLogFilter) ([]*repository.AuditLog, int, error) {
	query := `SELECT id, user_id, role, action, target, details, ip_address, user_agent, tenant_id, timestamp FROM infra_audit_logs WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", idx)
		args = append(args, *filter.Action)
		idx++
	}
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", idx)
		args = append(args, *filter.UserID)
		idx++
	}
	if filter.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", idx)
		args = append(args, *filter.TenantID)
		idx++
	}

	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*repository.AuditLog
	for rows.Next() {
		var log repository.AuditLog
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.Role, &log.Action, &log.Target, &log.Details,
			&log.IPAddress, &log.UserAgent, &log.TenantID, &log.Timestamp,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &log)
	}

	// total count
	countQuery := `SELECT COUNT(*) FROM infra_audit_logs WHERE 1=1`
	argsCount := []interface{}{}
	idx = 1
	if filter.Action != nil {
		countQuery += fmt.Sprintf(" AND action = $%d", idx)
		argsCount = append(argsCount, *filter.Action)
		idx++
	}
	if filter.UserID != nil {
		countQuery += fmt.Sprintf(" AND user_id = $%d", idx)
		argsCount = append(argsCount, *filter.UserID)
		idx++
	}
	if filter.TenantID != nil {
		countQuery += fmt.Sprintf(" AND tenant_id = $%d", idx)
		argsCount = append(argsCount, *filter.TenantID)
		idx++
	}

	var total int
	if err := r.DB.QueryRowContext(ctx, countQuery, argsCount...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
