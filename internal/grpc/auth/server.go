package authgrpc

import (
	"context"
	"errors"
	ssov1 "github.com/eclipsemode/go-grpc-sso-protobuf/gen/go/sso"
	"github.com/eclipsemode/go-grpc-sso/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emptyValue = 0
)

const (
	ErrMissingEmail           = "email is missing"
	ErrMissingPassword        = "password is missing"
	ErrMissingAppId           = "appId is missing"
	ErrMissingUserId          = "userId is missing"
	ErrInvalidEmailOrPassword = "invalid email or password"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		appID int) (token string, err error)
	RegisterNewUser(ctx context.Context,
		email string,
		password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrMissingEmail)
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrMissingPassword)
	}

	if req.GetAppId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, ErrMissingAppId)
	}

	// implement login via suite service
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, ErrInvalidEmailOrPassword)
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrMissingEmail)
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, ErrMissingPassword)
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, auth.ErrUserExists.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	if req.GetUserId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, ErrMissingUserId)
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, auth.ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}
