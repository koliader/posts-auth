package grpc_service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	db "github.com/koliader/posts-auth.git/internal/db/sqlc"
	"github.com/koliader/posts-auth.git/internal/pb"
	"github.com/koliader/posts-auth.git/internal/rabbitmq"
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

	err = s.redisClient.DeleteKey("users")
	if err != nil {
		return nil, errorResponse(codes.Internal, fmt.Sprintf("error to delete redis: %v", err))
	}

	// * Sign token
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
	redisUsers, err := s.redisClient.Get("users")
	// set users if redis is empty
	if err == redis.Nil {
		users, err := s.store.ListUsers(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unimplemented, "failed to list users")
		}
		convertedUsers := convertUsers(users)
		jsonStringUsers, err := json.Marshal(convertedUsers)
		if err != nil {
			return nil, errorResponse(codes.Internal, fmt.Sprintf("error to marshal users: %v", err))
		}
		err = s.redisClient.Set("users", jsonStringUsers)
		if err != nil {
			return nil, errorResponse(codes.Internal, fmt.Sprintf("error to set users into redis: %v", err))
		}
		res := &pb.ListUsersRes{
			Users: convertedUsers,
		}
		return res, nil
	}
	// unmarshal users
	var jsonUsers []db.User
	err = json.Unmarshal([]byte(*redisUsers), &jsonUsers)
	if err != nil {
		return nil, errorResponse(codes.Internal, fmt.Sprintf("error to unmarshal redis users: %v", err))
	}
	convertedUsers := convertUsers(jsonUsers)
	res := &pb.ListUsersRes{
		Users: convertedUsers,
	}
	return res, nil
}

func (s *Server) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailReq) (*pb.UserRes, error) {
	user, err := s.store.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorResponse(codes.NotFound, userNotFound)
		}
		return nil, errorResponse(codes.Unimplemented, "error to get user")
	}
	res := pb.UserRes{
		User: convertUser(user),
	}
	return &res, nil
}

func (s *Server) UpdateUserEmail(ctx context.Context, req *pb.UpdateUserEmailReq) (*pb.UserRes, error) {
	email, err := s.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}
	arg := db.UpdateUserEmailParams{
		Email:   *email,
		Email_2: req.NewEmail,
	}
	user, err := s.store.UpdateUserEmail(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorResponse(codes.NotFound, userNotFound)
		}
		return nil, errorResponse(codes.Unimplemented, "error to update user")
	}

	// res
	message := rabbitmq.UpdateEmailMessage{
		Email:    *email,
		NewEmail: req.NewEmail,
	}
	messageBody, err := json.Marshal(message)
	if err != nil {
		return nil, errorResponse(codes.Internal, fmt.Sprintf("failed to serialize message: %v", err))
	}

	err = s.rbmClient.SendMessage("updateUserEmail", []byte(messageBody))
	if err != nil {
		return nil, errorResponse(codes.Internal, fmt.Sprintf("error sending RabbitMQ message: %v", err))
	}

	err = s.redisClient.DeleteKey("users")
	if err != nil {
		return nil, errorResponse(codes.Internal, fmt.Sprintf("error to delete redis key: %v", err))
	}

	res := pb.UserRes{
		User: convertUser(user),
	}
	return &res, nil
}
