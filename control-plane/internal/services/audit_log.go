package services

import (
	"context"

	"google.golang.org/grpc/metadata"
)

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
