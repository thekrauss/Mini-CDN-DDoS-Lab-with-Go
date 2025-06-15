package handlers

// // gRPC handler → délègue à la logique métier
// func (h *NodeHandler) RegisterNode(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
// 	log.Println("→ RegisterNode handler déclenché")
// 	return h.Service.RegisterNode(ctx, req)
// }

// func (s *NodeHandler) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
// 	log.Printf("→ Ping from node %s: CPU=%.2f MEM=%.2f", req.NodeId, req.Cpu, req.Memory)
// 	return &pb.PingResponse{Status: "ok"}, nil
// }

// func (s *NodeHandler) SendMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.MetricsResponse, error) {
// 	log.Printf("→ Metrics from node %s: CPU=%.2f MEM=%.2f", req.NodeId, req.Cpu, req.Memory)
// 	return &pb.MetricsResponse{Status: "received"}, nil
// }
