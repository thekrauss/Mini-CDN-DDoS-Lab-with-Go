package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/handlers"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/middleware"

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erreur chargement config: %v", err)
	}
	config.AppConfig = *cfg

	if cfg.GCloudKeyPath != "" {
		_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.GCloudKeyPath)
		log.Println("GOOGLE_APPLICATION_CREDENTIALS défini sur :", cfg.GCloudKeyPath)
	}

	// Init DB
	store := &db.DBStore{}
	if _, err := store.OpenDatabase(cfg); err != nil {
		log.Fatalf("Échec connexion DB : %v", err)
	}
	defer store.CloseDatabase()

	if err := store.ApplyMigrations(); err != nil {
		log.Fatalf("Erreur migration DB: %v", err)
	}

	// Init Redis
	//	services.InitRedis(cfg)

	// Prometheus
	// if cfg.Metrics.PrometheusEnabled {
	// 	go func() {
	// 		log.Printf("Serveur Prometheus sur :%d", cfg.Metrics.PrometheusPort)
	// 		http.Handle("/metrics", services.NewPrometheusHandler())
	// 		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Metrics.PrometheusPort), nil))
	// 	}()
	// }

	// gRPC
	grpcAddr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Échec écoute gRPC: %v", err)
	}

	grpcServer := newGRPCServer(cfg)
	nodeServer := &handlers.NodeService{Store: store}
	pb.RegisterNodeServiceServer(grpcServer, nodeServer)

	go func() {
		log.Printf("Control-Plane gRPC sur %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Erreur lancement gRPC: %v", err)
		}
	}()

	// gRPC Gateway
	httpAddr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.HTTPPort)
	ctx := context.Background()
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Erreur client gRPC gateway: %v", err)
	}

	gwmux := runtime.NewServeMux()
	if err := pb.RegisterNodeServiceHandler(ctx, gwmux, conn); err != nil {
		log.Fatalf("Erreur enregistrement NodeService dans la gateway : %v", err)
	}

	gwServer := &http.Server{
		Addr:    httpAddr,
		Handler: gwmux,
	}

	go func() {
		log.Printf("API Gateway HTTP sur %s", httpAddr)
		if err := gwServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur serveur HTTP: %v", err)
		}
	}()

	waitForShutdown(grpcServer, gwServer)
}

func newGRPCServer(cfg *config.Config) *grpc.Server {
	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.LoggingMiddleware(),
			middleware.AuthMiddleware(cfg),
			middleware.RateLimitingMiddleware(),
			middleware.TimeoutMiddleware(),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			Timeout:               5 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
		}),
	)
}

func waitForShutdown(grpcServer *grpc.Server, gwServer *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Arrêt du control-plane...")

	grpcServer.GracefulStop()
	if err := gwServer.Close(); err != nil {
		log.Printf("Erreur arrêt HTTP: %v", err)
	}
}
