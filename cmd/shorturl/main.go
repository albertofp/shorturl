package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/albertofp/shorturl/api/proto/shorturl/v1"
	"github.com/albertofp/shorturl/api/server"
	"github.com/albertofp/shorturl/pkg/config"
	"github.com/albertofp/shorturl/pkg/db/sqlite"
)

func main() {
	cfg := config.New()
	l := slog.Default()

	sqliteDB, err := sqlite.New(cfg.DBPath)
	if err != nil {
		l.Error("failed to open database", slog.String("error", err.Error()))
		return
	}
	l.Info("Opened database", slog.String("path", cfg.DBPath))
	defer sqliteDB.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		l.Error("failed to listen", slog.String("error", err.Error()), slog.Int("port", cfg.GRPCPort))
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterURLShortenerServer(grpcServer, &server.URLShortenerServer{DB: sqliteDB})
	reflection.Register(grpcServer)

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		l.Info("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		os.Exit(0)
	}()

	l.Info("Starting gRPC server", slog.Int("port", cfg.GRPCPort))
	if err := grpcServer.Serve(listener); err != nil {
		l.Error("Failed to start gRPC server", slog.String("error", err.Error()))
		return
	}
}
