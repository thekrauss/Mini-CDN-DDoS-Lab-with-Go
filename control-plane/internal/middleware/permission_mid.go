package middleware

// import (
// 	"context"
// 	"log"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"

// 	authpb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
// 	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
// 	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/auth"
// )

// // Map des permissions requises par RPC
// var rpcPermissions = map[string]string{
// 	"/nodepb.NodeService.RegisterNode":         "MANAGE_NODE",
// 	"/nodepb.NodeService.Ping":                 "PING_NODE",
// 	"/nodepb.NodeService.GetAuditLogs":         "READ_AUDIT_LOGS",
// 	"/nodepb.NodeService.ListNodesByTenant":    "READ_NODE",
// 	"/nodepb.NodeService.UpdateNodeMetadata":   "MANAGE_NODE",
// 	"/nodepb.NodeService.SetNodeStatus":        "MANAGE_NODE",
// 	"/nodepb.NodeService.GetNodeByID":          "READ_NODE",
// 	"/nodepb.NodeService.CountActiveNodes":     "READ_NODE",
// 	"/nodepb.NodeService.ListBlacklistedNodes": "MANAGE_NODE",
// 	"/nodepb.NodeService.BlacklistNode":        "MANAGE_NODE",
// 	"/nodepb.NodeService.UnblacklistNode":      "MANAGE_NODE",
// 	"/nodepb.NodeService.SearchNodes":          "READ_NODE",
// 	"/nodepb.NodeService.StreamCommands":       "MANAGE_NODE",
// 	"/nodepb.NodeService.ReportCommandResult":  "MANAGE_NODE",
// }

// // Mapping des méthodes stream vers les permissions nécessaires
// var streamMethodPermissions = map[string]string{
// 	"/nodepb.NodeService.StreamCommands": "MANAGE_NODE",
// }

// // CheckPermissionInterceptor vérifie automatiquement la permission liée à chaque méthode RPC
// func CheckPermissionInterceptor(authClient authpb.AuthServiceClient) grpc.UnaryServerInterceptor {
// 	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 		permission, exists := rpcPermissions[info.FullMethod]
// 		if !exists {
// 			return handler(ctx, req) // aucune permission spécifique → continue
// 		}

// 		if err := verifyPermission(ctx, authClient, permission); err != nil {
// 			log.Printf("[PERM] %s refusée: %v", permission, err)
// 			return nil, err
// 		}

// 		return handler(ctx, req)
// 	}
// }

// // Intercepteur gRPC pour les streams
// func CheckPermissionStreamInterceptor(authClient authpb.AuthServiceClient) grpc.StreamServerInterceptor {
// 	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
// 		requiredPerm, found := streamMethodPermissions[info.FullMethod]
// 		if !found {
// 			return handler(srv, ss)
// 		}

// 		ctx := ss.Context()
// 		if err := verifyPermission(ctx, authClient, requiredPerm); err != nil {
// 			log.Printf("[PERM-STREAM] %s refusée: %v", requiredPerm, err)
// 			return status.Errorf(status.Code(err), err.Error())
// 		}

// 		return handler(srv, ss)
// 	}
// }

// // func partagée de vérification de permission
// func verifyPermission(ctx context.Context, authClient authpb.AuthServiceClient, requiredPerm string) error {
// 	claims, err := auth.ExtractJWTFromContext(ctx)
// 	if err != nil {
// 		return status.Errorf(codes.Unauthenticated, "Token requis ou invalide")
// 	}
// 	if config.IsSuperAdmin(claims.Role) {
// 		return nil
// 	}

// 	resp, err := authClient.HasPermission(ctx, &authpb.HasPermissionRequest{
// 		UserId:     claims.UserID.String(),
// 		Permission: requiredPerm,
// 	})
// 	if err != nil {
// 		return status.Errorf(codes.Internal, "Erreur lors de la vérification des permissions")
// 	}
// 	if !resp.Allowed {
// 		return status.Errorf(codes.PermissionDenied, "Permission refusée : %s", requiredPerm)
// 	}
// 	return nil
// }
