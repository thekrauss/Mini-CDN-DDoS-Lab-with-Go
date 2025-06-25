package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type AuditInput struct {
	UserID    string
	Role      string
	Action    string
	TargetID  string
	Details   string
	IPAddress string
	UserAgent string
	TenantID  string
	Timestamp time.Time
}

func AuditWorkflow(ctx workflow.Context, input AuditInput) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	})

	var resultErr error
	err := workflow.ExecuteActivity(ctx, "AuditActivity.LogAuditActivity", LogAuditInput{
		UserID:    input.UserID,
		Role:      input.Role,
		Action:    input.Action,
		Target:    input.TargetID,
		Details:   input.Details,
		IP:        input.IPAddress,
		UserAgent: input.UserAgent,
		TenantID:  input.TenantID,
	}).Get(ctx, &resultErr)

	return err
}
