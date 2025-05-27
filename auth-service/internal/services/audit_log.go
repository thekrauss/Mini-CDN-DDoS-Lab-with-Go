package services

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type AuditLog struct {
	ID         uuid.UUID
	AdminID    uuid.UUID  // ID de l'utilisateur ayant initié l'action
	Role       string     // Rôle de l'utilisateur (admin, opérateur, etc.)
	Action     string     // Action effectuée (ex: "Création utilisateur")
	TargetID   uuid.UUID  // ID de la ressource ciblée (ex: utilisateur affecté)
	TargetType string     // Type de la cible (ex: "Utilisateur", "Node")
	Details    string     // Détails de l'action
	IPAddress  string     // Adresse IP de l'appelant
	UserAgent  string     // Agent utilisateur (navigateur ou client HTTP)
	ActionTime time.Time  // Date et heure de l'action
	Status     string     // Statut de l'opération (succès, échec)
	SessionID  *uuid.UUID // ID de session si disponible (nullable)
	TenantID   *uuid.UUID // Identifiant d'organisation, entreprise ou équipe
}

// LogAction insère une entrée dans la table audit_logs
func (s *AuthService) LogAction(ctx context.Context, logEntry AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id_audit, admin_id, role, action, target_id, target_type,
			details, ip_address, user_agent, action_time, status, session_id, id_tenant
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	id := logEntry.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	_, err := s.Store.DB.ExecContext(ctx, query,
		id,
		logEntry.AdminID,
		logEntry.Role,
		logEntry.Action,
		logEntry.TargetID,
		logEntry.TargetType,
		logEntry.Details,
		logEntry.IPAddress,
		logEntry.UserAgent,
		logEntry.ActionTime,
		logEntry.Status,
		logEntry.SessionID,
		logEntry.TenantID,
	)

	if err != nil {
		log.Printf("Erreur log audit : %v", err)
		return err
	}

	log.Println("Action enregistrée dans les logs d’audit")
	return nil
}

func GetRequestMetadata(ctx context.Context) (string, string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "unknown", "unknown"
	}

	ip := "unknown"
	userAgent := "unknown"

	if values := md.Get("x-forwarded-for"); len(values) > 0 {
		ip = values[0]
	}

	if values := md.Get("user-agent"); len(values) > 0 {
		userAgent = values[0]
	}

	return ip, userAgent
}
