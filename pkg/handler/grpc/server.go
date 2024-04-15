package grpc_service

import (
	"fmt"

	db "github.com/koliader/posts-auth.git/internal/db/sqlc"
	"github.com/koliader/posts-auth.git/internal/pb"
	"github.com/koliader/posts-auth.git/internal/token"
	"github.com/koliader/posts-auth.git/internal/util"
)

type Server struct {
	pb.UnimplementedAuthServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create tokenManager: %v", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
