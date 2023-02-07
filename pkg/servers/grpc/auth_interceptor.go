package grpc

import (
	"context"

	"github.com/vitermakov/otusgo-final/pkg/servers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	authService servers.AuthService
}

func NewAuthInterceptor(authService servers.AuthService) *AuthInterceptor {
	return &AuthInterceptor{authService}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		user, err := i.authorize(ctx)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, servers.CtxKey{}, map[string]string{
			"id":    user.ID,
			"name":  user.Name,
			"login": user.Login,
		})
		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) authorize(ctx context.Context) (*servers.AuthUser, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}
	values := meta["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}
	userEmail := values[0]
	user, err := i.authService.Authorize(ctx, userEmail)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "error authorize: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "user e-mail is invalid: %s", userEmail)
	}
	return user, nil
}
