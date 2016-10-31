package auth

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/big"
	"time"

	"github.com/uber-go/zap"
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

	jwtIssuer = "accounts.google.com"

	// XXX: if google rotates keys, this will break. getting them programmatically is a
	// bit of a hassle since Google's APIs are dense. punt for now.
	googlePubkeys = map[string]*rsa.PublicKey{
		"f3ac035dfb99d1c6f12015c014555242317159df": googleKeyToRSAPublicKey("sBjVwZBJe3rjlGM2dc_qCahsrXTlXIk4Sk7wPRQi7l6CM5UYbd_kKu3fD_uCkLPnEhuylWfgf2a98SIvvOlNAfMG7hR82MN-YpgKrvJozQqcEsvQVPKOCaqK6xYeSA_Ag1EEP3UbOUZ4CxrSIS5-COXZv4cTR1hvZoEEHFaO1JG4NcTM7FsRcrjFBI9Smi9mD5YYOsLJqBLlB4vGpfRphxqmeHV_xmiLT2KpI2ArZaJODDkD503WUjHIkCjKwOoO2Pk0ZGciX8wvmn5mllzu-fj9-CRas-4IVLGVWs2clnUFPDLcwns8qNscDVLMZM8cuZM4RY4PirFZp0cXNxInWw==", "AQAB"),
		"9fb98df7486e2c58867c70485f8350334d12d977": googleKeyToRSAPublicKey("7BvEEbo8G0TbI9IbNpv6IKTg0lOBUgSYSvpNYHvrciiWovMdWNGtpBTlIQNcYVobGkOe1f67EVAFo8h3KBgeobSGvQMVTj_kRNCfhNVhq7hnsKLF17fAvyiprVXeC44IXdhHB4T03BInoj4xEqUo7QpBOtsi9GXkeu07YK-TBu0lAAuPjH503TWho6JulKIoD0KDj_dJtMp8OVrQAlvG2-SOQB9cg3k0SxG2G5sv0F5WWK02jBSqk7622Oqxle0Ur5yTDWYgpDJhxaSDEYiGVGMdHFKpD5f9JNla_1_r1SJCEW-t5XwywdhZSmimqGSSc-8wiJvX-Q9f8sW8-rE5ew==", "AQAB"),
		"73e336a6fc0a9501c30b4da8fe0388c666cad6b4": googleKeyToRSAPublicKey("lhIvlseHPwclH4SbRHBN6s7DJ__rt40RJvxqXhVJ3kwHdz_V0O8n93C_5eLHV6nWrUa6ezNR3fdVH7YS7HRkHU104bBEOU0vKh2O5B3tgS1wXymsySRsElcBDWnxPKkz58ASkNfOHX9mu_LtGUwAQ-YcnsCt42U_D9U6djb-GfRlKHHB1SC6OSza_kMOEYQC8pVWaYmyPq8JXxm-XR_VHlyg0wQMGpnrcWHBSmd52fVvi_7IoiuCczNil0tRH79CRR-YVZEekwlXkYzc_VQby0BIp5BpkI9BaAKFaKDIDW8fZ4ZUeIoW_d3nJRrtEjQHabzLQaJhkbDnrsNd_2ePNw==", "AQAB"),
		"04403cc231ae93a55fa157c4018b519a6a776282": googleKeyToRSAPublicKey("uEVCclCzlQ7xFwzVGQY8mDAHvF2PnQ4kHHFkzOf2WRiDFe87-adtuODQ5B83Bk7intm86McMkAbUzTvdpOmuZKZOfHe8HL8NGsg8M22hIZaeVKL0fz5v9_FyccmxVIdz230ReJNGLtz0uThurNAzi7y7pG3-3jGXq9vuSv9mzwwzzkWmyEqpRADn4VhgAMM9bAyk4Xp44kgJs0WeUNXTQocpmsEIhFW2ripDCn4v9Y0XZzXgrHwcshtiy-7qB87KFbSPctrKUrIhgRBJzjc8rwUYw7ZXDLRrZPX1kpbzK-q7uK-vXXXgVnCoqMJX16VlT-N8DkohgP-kHvk-8Dqi8w==", "AQAB"),
	}
)

// ClientConfig holds an oauth token (with refresh) and JWT. We have our own struct so we can serialize to json for on-disk conf.
type ClientConfig struct {
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

var _ credentials.PerRPCCredentials = &ClientJWT{}

func RefreshJWT(ctx context.Context, conf *ClientConfig, oauthConf *oauth2.Config, ll zap.Logger) (*ClientJWT, error) {
	_, err := validateToken(conf.JWT.Token, ll)
	if err != nil {
		if err != jwt.ErrExpired {
			return nil, err
		}

		// attempt to refresh with the oauth token
		ts := oauthConf.TokenSource(ctx, conf.OauthToken)
		token, err := ts.Token()
		if err != nil {
			return nil, err
		}
		return NewClientJWTFromOauth2(token, ll)
	}

	// token is good!
	return conf.JWT, nil
}

// NewClientJWTFromOauth2 creates a new ClientJWT from an oauth2 token.
func NewClientJWTFromOauth2(token *oauth2.Token, ll zap.Logger) (*ClientJWT, error) {
	jwtExtra := token.Extra("id_token")
	var jwtStr string
	var ok bool
	if jwtStr, ok = jwtExtra.(string); !ok || jwtStr == "" {
		ll.Warn("id_token field was empty or invalid")
		return nil, errInvalidOauth2Token
	}

	return &ClientJWT{Token: jwtStr}, nil
}

// googleKeyToRSAPublicKey converts a base64 publickey from Google's oauth cert list into an rsa.PublicKey or panics.
func googleKeyToRSAPublicKey(nstr, estr string) *rsa.PublicKey {
	decN, err := base64.URLEncoding.DecodeString(nstr)
	if err != nil {
		panic(err)
	}
	n := big.NewInt(0)
	n.SetBytes(decN)

	decE, err := base64.URLEncoding.DecodeString(estr)
	if err != nil {
		panic(err)
	}
	var eBytes []byte
	if len(decE) < 8 {
		eBytes = make([]byte, 8-len(decE), 8)
		eBytes = append(eBytes, decE...)
	} else {
		eBytes = decE
	}
	eReader := bytes.NewReader(eBytes)
	var e uint64
	err = binary.Read(eReader, binary.BigEndian, &e)
	if err != nil {
		panic(err)
	}
	return &rsa.PublicKey{N: n, E: int(e)}
}

func validateToken(token string, ll zap.Logger) (*jwt.JSONWebToken, error) {
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
		return nil, errUnknownSigningKey
	}

	var sharedKey *rsa.PublicKey
	var ok bool
	if sharedKey, ok = googlePubkeys[keyID]; !ok {
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
