package grpc_service

import (
	"context"

	db "github.com/koliader/posts-auth.git/internal/db/sqlc"
	"github.com/koliader/posts-auth.git/internal/pb"
	"github.com/koliader/posts-auth.git/internal/util"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Register(ctx context.Context, req *pb.RegisterReq) (*pb.AuthRes, error) {
	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	arg := db.CreateUserParams{
		Email:    req.GetEmail(),
		Username: req.GetUsername(),
		Password: hashedPassword,
	}
	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, errorResponse(codes.AlreadyExists, "username already exists")
			}
		}
		return nil, status.Errorf(codes.Unimplemented, "failed to create user")
	}

	token, err := s.tokenMaker.CreateToken(user.Email, s.config.AccessTokenDuration)
	if err != nil {
		return nil, errorResponse(codes.Internal, "error creating token")
	}

	res := &pb.AuthRes{
		Token: token,
	}
	return res, nil
}

func (s *Server) ListUsers(ctx context.Context, req *pb.Empty) (*pb.ListUsersRes, error) {
	var convertedUsers []*pb.UserEntity
	users, err := s.store.ListUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unimplemented, "failed to list users")
	}
	for _, user := range users {
		convertedUsers = append(convertedUsers, convertUser(user))
	}
	res := &pb.ListUsersRes{
		Users: convertedUsers,
	}
	return res, nil
}
