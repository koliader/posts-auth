package main

import (
	"context"
	"fmt"
	"net"
	"os"

	// "log"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/koliader/posts-auth.git/internal/db/sqlc"
	"github.com/koliader/posts-auth.git/internal/pb"
	"github.com/koliader/posts-auth.git/internal/util"
	grpc_service "github.com/koliader/posts-auth.git/pkg/handler/grpc"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("cannot load config")
	}
	if config.Environment == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Msg("cannot connect to db")
	}

	defer connPool.Close()

	store := db.NewStore(connPool)
	runGrpcServer(config, store)
}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := grpc_service.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("cannot create server: %v", err))
	}
	defer server.Close() // Закрыть соединение с RabbitMQ при завершении работы сервера

	listener, err := net.Listen("tcp", config.ServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create listener")
	}

	// create logger
	grpcLogger := grpc.UnaryInterceptor(grpc_service.GrpcLogger)
	// create server
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterAuthServer(grpcServer, server)
	reflection.Register(grpcServer)

	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Msg("cannot start gRPC server")
	}
}
