package main

import (
	"context"
	"errors"
	"film-searcher/internal/adapter"
	"film-searcher/internal/adapter/postgres"
	"film-searcher/internal/config"
	gh "film-searcher/internal/handler/grpc"
	hh "film-searcher/internal/handler/http"
	"film-searcher/internal/usecase"
	"film-searcher/migrations"
	pb "film-searcher/proto"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
)

const timeout = 30 * time.Second

func run() error {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	migrationDB := stdlib.OpenDBFromPool(pool)
	defer migrationDB.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, migrationDB, migrations.Files)
	if err != nil {
		return fmt.Errorf("create goose: %w", err)

	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	repo := postgres.New(pool)
	stand := adapter.New(cfg.BaseURL, timeout)
	uc := usecase.New(stand, repo)
	if err := uc.Sync(ctx); err != nil {
		return fmt.Errorf("sync movies: %w", err)
	}

	httpHandler := hh.New(uc, timeout)
	grpcHandler := gh.New(uc)

	mux := http.NewServeMux()

	httpHandler.RegisterRoutes(mux)
	mux.Handle("/swagger/", hh.SwaggerHandler())
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
		log.Printf("Swagger UI: http://localhost%s/swagger/index.html", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("failed to listen grpc: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMovieServiceServer(grpcServer, grpcHandler)

	go func() {
		log.Printf("gRPC server listening on %s", cfg.GRPCAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()

	fmt.Println("server stopped")

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
