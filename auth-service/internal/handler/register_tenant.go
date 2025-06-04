package handler

import (
	"context"
	"log"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/pkg/auth"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AuthServer) CreateTenant(ctx context.Context, req *pb.CreateTenantRequest) (*pb.CreateTenantResponse, error) {
	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		log.Printf("Erreur JWT: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}
	return s.Service.CreateTenant(ctx, req, claims)
}

func (s *AuthServer) DeleteTenant(ctx context.Context, req *pb.DeleteTenantRequest) (*emptypb.Empty, error) {
	claims, err := auth.ExtractJWTFromContext(ctx)
	if err != nil {
		log.Printf("Erreur JWT: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou expiré")
	}
	return s.Service.DeleteTenant(ctx, req, claims)
}
