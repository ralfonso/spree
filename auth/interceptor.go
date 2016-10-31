package auth

import (
	"crypto/rsa"
	"strings"

	"github.com/uber-go/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"gopkg.in/square/go-jose.v2/jwt"
)

func isAuthorizedToken(tok *jwt.JSONWebToken, allowedEmails []string, ll zap.Logger) (bool, error) {
	headers := tok.Headers
	var keyID string
	for _, header := range headers {
		if header.KeyID != "" {
			keyID = header.KeyID
			break
		}
	}
	if keyID == "" {
		return false, errUnknownSigningKey
	}

	var sharedKey *rsa.PublicKey
	var ok bool
	if sharedKey, ok = googlePubkeys[keyID]; !ok {
		return false, errUnknownSigningKey
	}

	emailClaims := struct {
		Email string `json:"email"`
	}{}
	if err := tok.Claims(sharedKey, &emailClaims); err != nil {
		return false, err
	}

	var authorized bool
	for _, checkEmail := range allowedEmails {
		if strings.ToLower(checkEmail) == strings.ToLower(emailClaims.Email) {
			authorized = true
			break
		}
	}

	if !authorized {
		return false, errUnauthorizedEmail
	}

	return true, nil
}

// MakeJWTInterceptor creates an interceptor to validate JWT tokens for a unary RPC.
func MakeJWTInterceptor(allowedEmails []string, ll zap.Logger) grpc.UnaryServerInterceptor {
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

		tok, err := validateToken(jwtTokenStr[0], ll)
		if err != nil {
			ll.Warn("invalid token in RPC", zap.Object("req", req), zap.Error(err))
			return nil, grpc.Errorf(codes.Unauthenticated, "valid token required.")
		}

		authorized, err := isAuthorizedToken(tok, allowedEmails, ll)
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
