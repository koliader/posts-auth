package grpc_service

import (
	"context"

	"github.com/jackc/pgx/v5"
	db "github.com/koliader/posts-auth.git/internal/db/sqlc"
	"github.com/koliader/posts-auth.git/internal/pb"
	"github.com/koliader/posts-auth.git/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	userNotFound = "user not found"
	authError    = "invalid login or password"
)

// * Register

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
		if code := db.ErrorCode(err); code == db.UniqueViolation {
			return nil, status.Errorf(codes.AlreadyExists, "user with this email is created")
		}
		return nil, status.Errorf(codes.Unimplemented, "failed to create user: %v", err)
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

// * Login
func (s *Server) Login(ctx context.Context, req *pb.LoginReq) (*pb.AuthRes, error) {
	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.Unauthenticated, authError)
		}
		return nil, status.Errorf(codes.Unimplemented, "error to get user")
	}
	passwordIsEqual := util.CheckPassword(user.Password, req.Password)
	if passwordIsEqual != nil {
		return nil, status.Errorf(codes.Unauthenticated, authError)
	}

	token, err := s.tokenMaker.CreateToken(user.Email, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Error(codes.Unimplemented, "error to sign token")
	}
	res := pb.AuthRes{
		Token: token,
	}
	return &res, nil
}

// * List

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
