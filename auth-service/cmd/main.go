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

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/config"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/db"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/handler"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/middleware"
	"github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/internal/services"

	pb "github.com/thekrauss/Mini-CDN-DDoS-Lab-with-Go/auth-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Échec du chargement de la configuration : %v", err)
	}
	config.AppConfig = *cfg

	if cfg.GCloudKeyPath != "" {
		_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.GCloudKeyPath)
		log.Println("GOOGLE_APPLICATION_CREDENTIALS défini sur :", cfg.GCloudKeyPath)
	}

	store := &db.DBStore{}
	if _, err := store.OpenDatabase(cfg); err != nil {
		log.Fatalf("Échec connexion DB : %v", err)
	}
	defer store.CloseDatabase()

	if err := store.ApplyMigrations(); err != nil {
		log.Fatalf("Échec des migrations : %v", err)
	}

	services.InitRedis(cfg)

	if _, err = services.InitFirebase(); err != nil {
		log.Fatalf("Erreur init Firebase : %v", err)
	}
	log.Println("Firebase initialisé ")

	grpcAddr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Échec écoute sur %s : %v", grpcAddr, err)
	}

	grpcServer := initializeGRPCServer(cfg)
	authServer := &handler.AuthServer{Store: store}
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	go func() {
		log.Printf("Auth service running on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	gatewayAddr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.HTTPPort)
	ctx := context.Background()
	conn, err := grpc.NewClient(gatewayAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Échec connexion gRPC client : %v", err)
	}

	gwmux := runtime.NewServeMux()

	if err := pb.RegisterAuthServiceHandler(ctx, gwmux, conn); err != nil {
		log.Fatalf("Échec enregistrement AuthService Gateway : %v", err)
	}

	gwServer := &http.Server{
		Addr: gatewayAddr,
	}

	log.Printf("API Gateway running on %s", gatewayAddr)
	if err := gwServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to serve API Gateway: %v", err)
	}

	defer os.Remove(config.AppConfig.GCloudKeyPath)

	waitForShutdown(grpcServer, gwServer)
}

func waitForShutdown(grpcServer *grpc.Server, gwServer *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Arrêt des serveurs...")

	grpcServer.GracefulStop()
	log.Println("Serveur gRPC arrêté")

	if err := gwServer.Close(); err != nil {
		log.Printf("Erreur arrêt HTTP : %v", err)
	} else {
		log.Println(" Serveur HTTP arrêté")
	}
}

func initializeGRPCServer(cfg *config.Config) *grpc.Server {
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
