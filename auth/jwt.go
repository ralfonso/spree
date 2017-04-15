package auth

import (
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/credentials"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	errUnknownSigningKey  = errors.New("unknown jwt signing key")
	errInvalidJWTToken    = errors.New("invalid jwt token")
	errInvalidOauth2Token = errors.New("invalid oauth2 token")
	errInvalidIssuer      = errors.New("invalid jwt issuer")
	errMissingEmail       = errors.New("jwt token is missing email")
	errUnauthorizedEmail  = errors.New("jwt token email is for unauthorized user")
)

const (
	jwtIssuer = "accounts.google.com"
	certUrl   = "https://www.googleapis.com/oauth2/v2/certs"
)

// ClientConfig holds an oauth token (with refresh) and JWT. We have our own struct so we can serialize to json for on-disk conf.
type ClientConfig struct {
	RPCAddr    string        `json:"rpc_addr"`
	OauthToken *oauth2.Token `json:"oauth_token"`
	JWT        *ClientJWT    `json:"jwt"`
}

// ClientJWT holds a JWT and implements credentials.PerRPCCredentials
type ClientJWT struct {
	Token string `json:"token"`
}

func (j *ClientJWT) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": j.Token,
	}, nil
}

func (j *ClientJWT) RequireTransportSecurity() bool {
	return true
}

// NewClientJWTFromOauth2 creates a new ClientJWT from an oauth2 token.
func NewClientJWTFromOauth2(token *oauth2.Token, ll *zap.Logger) (*ClientJWT, error) {
	jwtExtra := token.Extra("id_token")
	var jwtStr string
	var ok bool
	if jwtStr, ok = jwtExtra.(string); !ok || jwtStr == "" {
		ll.Warn("id_token field was empty or invalid")
		return nil, errInvalidOauth2Token
	}

	return &ClientJWT{Token: jwtStr}, nil
}

var _ credentials.PerRPCCredentials = &ClientJWT{}

// Authenticator can validate or refresh JWT tokens.
type Authenticator struct {
	keys keyCache
	ll   *zap.Logger
}

// NewAuthenticator returns a new Authenticator.
func NewAuthenticator(ll *zap.Logger) *Authenticator {
	return &Authenticator{
		keys: newPubkeyCache(certUrl, ll),
		ll:   ll,
	}
}

func (a *Authenticator) RefreshJWT(ctx context.Context, conf *ClientConfig, oauthConf *oauth2.Config) (*oauth2.Token, *ClientJWT, error) {
	_, err := a.ValidateToken(conf.JWT.Token)
	if err != nil {
		// attempt to refresh with the oauth token
		ts := oauthConf.TokenSource(ctx, conf.OauthToken)
		token, err := ts.Token()
		if err != nil {
			return nil, nil, err
		}
		jwt, err := NewClientJWTFromOauth2(token, a.ll)
		return token, jwt, err
	}

	// token is good!
	return conf.OauthToken, conf.JWT, nil
}

func (a *Authenticator) ValidateToken(token string) (*jwt.JSONWebToken, error) {
	tok, err := jwt.ParseSigned(token)
	if err != nil {
		return nil, err
	}

	headers := tok.Headers
	var keyID string
	for _, header := range headers {
		if header.KeyID != "" {
			keyID = header.KeyID
			break
		}
	}
	if keyID == "" {
		a.ll.Error("no signing key id")
		return nil, errUnknownSigningKey
	}

	sharedKey := a.keys.Get(keyID)
	if sharedKey == nil {
		a.ll.Error("unknown signing key id", zap.String("key.id", keyID))
		return nil, errUnknownSigningKey
	}

	cl := jwt.Claims{}
	if err := tok.Claims(sharedKey, &cl); err != nil {
		return nil, err
	}

	err = cl.Validate(jwt.Expected{
		Time:   time.Now(),
		Issuer: jwtIssuer,
	})
	if err != nil {
		return nil, err
	}

	return tok, nil
}

func (a *Authenticator) IsAuthorizedToken(tok *jwt.JSONWebToken, allowedEmails []string) (bool, error) {
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

	sharedKey := a.keys.Get(keyID)
	if sharedKey == nil {
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
