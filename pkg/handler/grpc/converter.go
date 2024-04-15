package grpc_service

import (
	db "github.com/koliader/posts-auth.git/internal/db/sqlc"
	"github.com/koliader/posts-auth.git/internal/pb"
)

func convertUser(user db.User) *pb.UserEntity {
	return &pb.UserEntity{
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	}
}
