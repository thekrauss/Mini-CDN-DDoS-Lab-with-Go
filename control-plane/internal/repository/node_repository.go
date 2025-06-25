package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

//
// ─── STRUCTURES PRINCIPALES ─────────────────────────────────────────────────────
//

type Node struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`             // Nom logique (ex: "cdn-eu-west-1")
	IP              string    `json:"ip"`               // IP publique ou privée
	TenantID        string    `json:"tenant_id"`        // Multi-tenant
	Status          string    `json:"status"`           // online, offline, degraded
	LastSeen        time.Time `json:"last_seen"`        // Dernier heartbeat
	CreatedAt       time.Time `json:"created_at"`       // Date création
	UpdatedAt       time.Time `json:"updated_at"`       // Dernière mise à jour
	Location        string    `json:"location"`         // Ville ou zone (ex: Paris)
	Provider        string    `json:"provider"`         // aws, ovh, gcp, etc.
	SoftwareVersion string    `json:"software_version"` // Version de l'agent
	IsBlacklisted   bool      `json:"is_blacklisted"`   // Exclu temporairement ?
	Tags            []string  `json:"tags"`             // Pour UI, filtrage
	OS              string    `json:"os"`               // OS du node
}

type NodeConfig struct {
	NodeID          string            `json:"node_id"`          // identifiant du nœud
	PingInterval    int               `json:"ping_interval"`    // fréquence des ping() en secondes
	MetricsInterval int               `json:"metrics_interval"` // fréquence des sendMetrics() en secondes
	DynamicConfig   bool              `json:"dynamic_config"`   // true si config récupérée dynamiquement
	CustomLabels    map[string]string `json:"custom_labels"`    // paires clé/valeur pour tags ou méta
}

type NodeMetrics struct {
	NodeID      string    `json:"node_id"`
	TenantID    string    `json:"tenant_id"`
	Timestamp   time.Time `json:"timestamp"`
	CPU         float64   `json:"cpu"`
	Memory      float64   `json:"memory"`
	BandwidthRx int64     `json:"bandwidth_rx"`
	BandwidthTx int64     `json:"bandwidth_tx"`
	Connections int       `json:"connections"`
	DiskIO      int64     `json:"disk_io"`
	Uptime      int64     `json:"uptime"`
	Status      string    `json:"status"`
}

type NodeStatus string

const (
	NodeOnline   NodeStatus = "online"
	NodeOffline  NodeStatus = "offline"
	NodeDegraded NodeStatus = "degraded"
)

type NodeFilter struct {
	TenantID string
	Query    string
	Status   *NodeStatus
	TagKey   string
	TagValue string
	IP       string
}

//
// ─── UTILISATEUR (REDIS) ─────────────────────────────────────────────────────────
//

type UtilisateurRedis struct {
	IDUtilisateur string
	Nom           string
	Prenom        string
	Email         string
	Telephone     string
	Role          string
	Permissions   string
	TenantID      string
	MFAEnabled    bool
	IsActive      bool
	Status        string
}

//
// ─── AUDIT ───────────────────────────────────────────────────────────────────────
//

type AuditLog struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Role      string
	Action    string
	Target    string
	Details   string
	IPAddress string
	UserAgent string
	Timestamp time.Time
	TenantID  uuid.UUID
}

type AuditLogFilter struct {
	Limit    int
	Offset   int
	Action   *string
	UserID   *string
	TenantID *string
}

type NodeRepository interface {
	// CRUD
	CreateNode(ctx context.Context, node *Node) error
	GetNodeByID(ctx context.Context, id string) (*Node, error)
	UpdateHeartbeat(ctx context.Context, id string, seenAt time.Time) error
	ListNodesByTenant(ctx context.Context, tenantID string) ([]*Node, error)
	DeleteNode(ctx context.Context, id string) error

	// Node Config
	UpdateNodeConfig(ctx context.Context, node *NodeConfig) error
	DeleteNodeConfig(ctx context.Context, nodeID string) error
	GetNodeConfig(ctx context.Context, nodeID string) (*NodeConfig, error)

	// Recherche / Filtrage
	UpdateNodeMetadata(ctx context.Context, id string, name string, ip string, tags map[string]string) error
	SearchNodes(ctx context.Context, filter NodeFilter) ([]*Node, error)
	CountActiveNodes(ctx context.Context, tenantID string, since time.Duration) (int, error)

	// Statut / Orchestration
	SetNodeStatus(ctx context.Context, id string, status string) error
	SetNodeBlacklistStatus(ctx context.Context, nodeID string, isBlacklisted bool) error
	ListBlacklistedNodes(ctx context.Context, tenantID string) ([]*Node, error)
	GetInactiveNodes(ctx context.Context, olderThan time.Duration) ([]*Node, error)
	MarkAllNodesOffline(ctx context.Context) error

	// Sécurité / Multi-tenant
	IsIPAlreadyRegistered(ctx context.Context, ip string) (bool, error)
	AssignToTenant(ctx context.Context, nodeID string, tenantID string) error

	// Audit / Logs
	InsertAuditLog(ctx context.Context, log *AuditLog) error
	GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]*AuditLog, int, error)

	// Métriques
	StoreNodeMetrics(ctx context.Context, metrics *NodeMetrics) error
}
