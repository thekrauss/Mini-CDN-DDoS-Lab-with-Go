package services

import (
	"context"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/repository"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *NodeService) GetAuditLogs(ctx context.Context, req *pb.GetAuditLogsRequest) (*pb.GetAuditLogsResponse, error) {
	filter := repository.AuditLogFilter{
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	if req.Action != nil {
		filter.Action = req.Action
	}
	if req.UserId != nil {
		filter.UserID = req.UserId
	}
	if req.TenantId != nil {
		filter.TenantID = req.TenantId
	}

	logs, total, err := s.Repo.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur récupération logs : %v", err)
	}

	var entries []*pb.AuditLogEntry
	for _, log := range logs {
		entries = append(entries, &pb.AuditLogEntry{
			Id:        log.ID.String(),
			UserId:    log.UserID.String(),
			Role:      log.Role,
			Action:    log.Action,
			Target:    log.Target,
			Details:   log.Details,
			IpAddress: log.IPAddress,
			UserAgent: log.UserAgent,
			TenantId:  log.TenantID.String(),
			Timestamp: log.Timestamp.Format(time.RFC3339),
		})
	}

	return &pb.GetAuditLogsResponse{
		Logs:  entries,
		Total: int32(total),
	}, nil
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
