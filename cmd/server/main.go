package main

import (
	"context"
	"errors"
	"film-searcher/internal/adapter"
	"film-searcher/internal/config"
	gh "film-searcher/internal/handler/grpc"
	hh "film-searcher/internal/handler/http"
	"film-searcher/internal/usecase"
	pb "film-searcher/proto"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	timeout := 30 * time.Second

	client := adapter.New(cfg.BaseURL, timeout)
	uc := usecase.New(client)
	httpHandler := hh.New(uc, timeout)
	grpcHandler := gh.New(uc)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
}
