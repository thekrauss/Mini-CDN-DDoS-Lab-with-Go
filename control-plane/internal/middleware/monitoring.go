package middleware

import (
	"context"
	"time"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/monitoring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func PrometheusMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()

		code := "OK"
		if err != nil {
			code = status.Code(err).String()
		}

		monitoring.GRPCRequests.WithLabelValues(info.FullMethod, code).Inc()
		monitoring.RequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

		return resp, err
	}
}
