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

	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/monitoring"
	pkg "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/pkg/redis"
	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/proto"
	"go.temporal.io/sdk/client"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/config"
	workers "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/flush"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/middleware"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/services"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/control-plane/internal/ws"

	authpb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
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

	// init DB
	store := &db.DBStore{}
	if _, err := store.OpenDatabase(cfg); err != nil {
		log.Fatalf("Échec connexion DB : %v", err)
	}
	defer store.CloseDatabase()

	if err := store.ApplyMigrations(); err != nil {
		log.Fatalf("Erreur migration DB: %v", err)
	}

	//init Redis & Prometheus
	pkg.InitRedis(cfg)
	monitoring.Init()

	// Prometheus
	if cfg.Metrics.PrometheusEnabled {
		go func() {
			addr := ":" + strconv.Itoa(cfg.Metrics.PrometheusPort)
			log.Printf("Serveur Prometheus /metrics sur %s", addr)
			http.Handle("/metrics", monitoring.Handler())
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Fatalf("Erreur serveur Prometheus: %v", err)
			}
		}()
	}

	// Initialisation client Temporal
	var temporalClient client.Client
	if cfg.Temporal.Enabled {
		temporalClient, err = client.NewClient(client.Options{
			HostPort:  cfg.Temporal.Address,
			Namespace: cfg.Temporal.Namespace,
		})
		if err != nil {
			log.Fatalf("Erreur connexion Temporal: %v", err)
		}
		log.Println("Client Temporal initialisé")
		defer temporalClient.Close()
	}

	// gRPC
	grpcAddr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Échec écoute gRPC: %v", err)
	}

	authClient, err := NewAuthServiceClient(cfg.AuthService.Host)
	if err != nil {
		log.Fatalf("Erreur connexion auth-service: %v", err)
	}

	hub := ws.NewHub()

	grpcServer := newGRPCServer(cfg)
	nodeServer := &services.NodeService{
		Store:          store,
		AuthClient:     authClient,
		Hub:            hub,
		TemporalClient: temporalClient,
	}
	pb.RegisterNodeServiceServer(grpcServer, nodeServer)

	go func() {
		log.Printf("Control-Plane gRPC sur %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Erreur lancement gRPC: %v", err)
		}
	}()

	go func() {
		log.Println("Serveur WebSocket sur :8088/ws/nodes")
		http.Handle("/ws/nodes", ws.NewWSServer(hub))
		log.Fatal(http.ListenAndServe(":8088", nil))
	}()

	// flush périodique des heartbeat en DB
	repo := db.NewNodeRepository(store.DB)
	workers.StartPingFlushWorker(repo)
	workers.StartMetricsFlushWorker(repo)

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
			//middleware.CheckPermissionInterceptor(authClient),
			middleware.AuthMiddleware(cfg),
			middleware.RateLimitingMiddleware(),
			middleware.TimeoutMiddleware(),
			middleware.PrometheusMiddleware(),
		),
		// grpc.ChainStreamInterceptor(
		// 	middleware.CheckPermissionStreamInterceptor(authClient),
		// ),
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := gwServer.Shutdown(ctx); err != nil {
		log.Printf("Erreur arrêt HTTP: %v", err)
	}
}

func NewAuthServiceClient(addr string) (authpb.AuthServiceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // attend que la connexion soit prête
	)
	if err != nil {
		return nil, err
	}

	return authpb.NewAuthServiceClient(conn), nil
}
