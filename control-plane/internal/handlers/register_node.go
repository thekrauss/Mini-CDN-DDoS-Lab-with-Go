package handlers

import (
	"context"
	"log"

	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
)

func (s *NodeService) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	log.Printf("→ Ping from node %s: CPU=%.2f MEM=%.2f", req.NodeId, req.Cpu, req.Memory)
	return &pb.PingResponse{Status: "ok"}, nil
}

func (s *NodeService) SendMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsResponse, error) {
	log.Printf("→ Metrics from node %s: CPU=%.2f MEM=%.2f", req.NodeId, req.Cpu, req.Memory)
	return &pb.MetricsResponse{Status: "received"}, nil
}
