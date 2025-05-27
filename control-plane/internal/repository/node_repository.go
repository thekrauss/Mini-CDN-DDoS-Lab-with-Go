package repository

import (
	"context"
	"time"
)

type Node struct {
	ID        string
	Name      string
	IP        string
	TenantID  string
	Status    string
	LastSeen  time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
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
}
