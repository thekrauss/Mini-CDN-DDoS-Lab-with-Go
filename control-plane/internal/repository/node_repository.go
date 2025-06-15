package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Node struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`             //  (ex: "cdn-eu-west-1")
	IP              string    `json:"ip"`               //  IP publique ou interne du nœud
	TenantID        string    `json:"tenant_id"`        // ID du client propriétaire (multi-tenant)
	Status          string    `json:"status"`           // statut du node (alive, unreachable, disabled)
	LastSeen        time.Time `json:"last_seen"`        // date du dernier heartbeat reçu
	CreatedAt       time.Time `json:"created_at"`       // date d'enregistrement du nœud
	UpdatedAt       time.Time `json:"updated_at"`       // date de dernière mise à jour
	Location        string    `json:"location"`         // (ex: "Paris", "eu-west-1")
	Provider        string    `json:"provider"`         //  (aws, gcp, ovh, on-prem)
	SoftwareVersion string    `json:"software_version"` // version de l'agent exécuté sur le nœud
	IsBlacklisted   bool      `json:"is_blacklisted"`   //  si le nœud est temporairement désactivé (DDoS, infra)
	Tags            []string  `json:"tags"`             // mots-clés libres pour filtrage, UI, regroupement logique
	OS              string    `json:"os"`
}

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

type NodeMetrics struct {
	NodeID    string    `json:"node_id"`
	TenantID  string    `json:"tenant_id"`
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Uptime    int64     `json:"uptime"`
	Status    string    `json:"status"`
}

type NodeFilter struct {
	TenantID string
	Query    string
	Status   *NodeStatus
	TagKey   string
	TagValue string
	IP       string
}

type NodeStatus string

const (
	NodeOnline   NodeStatus = "online"
	NodeOffline  NodeStatus = "offline"
	NodeDegraded NodeStatus = "degraded"
)

type NodeRepository interface {
	// CRUD de base
	CreateNode(ctx context.Context, node *Node) error
	GetNodeByID(ctx context.Context, id string) (*Node, error)
	UpdateHeartbeat(ctx context.Context, id string, seenAt time.Time) error
	ListNodesByTenant(ctx context.Context, tenantID string) ([]*Node, error)
	DeleteNode(ctx context.Context, id string) error

	// Fonctions avancées
	UpdateNodeMetadata(ctx context.Context, id string, name string, ip string, tags map[string]string) error //renommer, changer IP, ou tags
	SearchNodes(ctx context.Context, filter NodeFilter) ([]*Node, error)                                     //filtres pour l’interface admin (status, IP, nom, tag…)
	CountActiveNodes(ctx context.Context, tenantID string, since time.Duration) (int, error)                 //pour usage SaaS : quotas, stats

	// Statut / Orchestration
	SetNodeStatus(ctx context.Context, id string, status string) error //ajoute de statut online, degraded, offline

	GetInactiveNodes(ctx context.Context, olderThan time.Duration) ([]*Node, error) //détection automatique des nœuds morts
	MarkAllNodesOffline(ctx context.Context) error                                  //Réinitialisation périodique

	// Sécurité / Enregistrement
	IsIPAlreadyRegistered(ctx context.Context, ip string) (bool, error)       // limite les enregistrements
	AssignToTenant(ctx context.Context, nodeID string, tenantID string) error //migration d’un node à un client

	InsertAuditLog(ctx context.Context, log *AuditLog) error
	GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]*AuditLog, int, error)
	StoreNodeMetrics(ctx context.Context, metrics *NodeMetrics) error
}
