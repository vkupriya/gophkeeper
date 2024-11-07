package interceptors

import (
	"context"

	"github.com/vkupriya/gophkeeper/internal/server/helpers"
	"github.com/vkupriya/gophkeeper/internal/server/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var userServiceLogin = "/proto.GophKeeper/Login"
var userServiceRegister = "/proto.GophKeeper/Register"
var ignoreMethod = []string{userServiceLogin, userServiceRegister}

func AuthInterceptor(cfg *models.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		for _, imethod := range ignoreMethod {
			if info.FullMethod == imethod {
				return handler(ctx, req)
			}
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		accessToken := values[0]
		claims, err := helpers.ValidateJWT(cfg, accessToken)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
		}

		md.Append("userid", claims.UserID)
		ctx = metadata.NewIncomingContext(ctx, md)
		return handler(ctx, req)
	}
}
