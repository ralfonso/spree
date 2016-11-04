package auth

import (
	"github.com/uber-go/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// MakeJWTInterceptor creates an interceptor to validate JWT tokens for a unary RPC.
func MakeJWTInterceptor(allowedEmails []string, authenticator *Authenticator, ll zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		md, ok := metadata.FromContext(ctx)
		if !ok {
			ll.Warn("missing metadata in RPC", zap.Object("req", req))
			return nil, grpc.Errorf(codes.Unauthenticated, "valid token required")
		}

		jwtTokenStr, ok := md["authorization"]
		if !ok {
			ll.Warn("missing authorization in RPC", zap.Object("req", req))
			return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
		}

		tok, err := authenticator.ValidateToken(jwtTokenStr[0])
		if err != nil {
			ll.Warn("invalid token in RPC", zap.Object("req", req), zap.Error(err))
			return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
		}

		authorized, err := authenticator.IsAuthorizedToken(tok, allowedEmails)
		if err != nil {
			ll.Warn("unauthorized token in RPC", zap.Object("req", req), zap.Error(err))
			return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
		}

		if !authorized {
			return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
		}

		return handler(ctx, req)
	}
}
