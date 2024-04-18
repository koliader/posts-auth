package grpc_service

import (
	"fmt"

	db "github.com/koliader/posts-auth.git/internal/db/sqlc"
	"github.com/koliader/posts-auth.git/internal/pb"
	"github.com/koliader/posts-auth.git/internal/rabbitmq"
	"github.com/koliader/posts-auth.git/internal/token"
	"github.com/koliader/posts-auth.git/internal/util"
)

type Server struct {
	pb.UnimplementedAuthServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	rbmClient  rabbitmq.Client
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create tokenManager: %v", err)
	}

	rbmClient, err := rabbitmq.NewClient(config, "updateUserEmail")
	if err != nil {
		return nil, fmt.Errorf("error creating rabbitmq client: %v", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		rbmClient:  *rbmClient,
	}

	return server, nil
}

// Close закрывает соединение с RabbitMQ
func (s *Server) Close() error {
	if err := s.rbmClient.Close(); err != nil {
		return err
	}
	return nil
}
