package workflows

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
)

type LogAuditInput struct {
	UserID    string
	Role      string
	Action    string
	Target    string
	Details   string
	IP        string
	UserAgent string
	TenantID  string
}

type AuditActivity struct {
	Repo repository.NodeRepository
}

func (a *AuditActivity) LogAuditActivity(ctx context.Context, input LogAuditInput) error {
	entry := &repository.AuditLog{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(input.UserID),
		Role:      input.Role,
		Action:    input.Action,
		Target:    input.Target,
		Details:   input.Details,
		IPAddress: input.IP,
		UserAgent: input.UserAgent,
		TenantID:  uuid.MustParse(input.TenantID),
		Timestamp: time.Now(),
	}

	return a.Repo.InsertAuditLog(ctx, entry)
}
