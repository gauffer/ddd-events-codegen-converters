package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/application/port"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var _ port.JWTSigner = (*JWTSigner)(nil)

type JWTSigner struct {
	pk     *ecdsa.PrivateKey
	issuer string
}

func NewJWTSigner(pkRawB64 string, issuer string) (*JWTSigner, error) {
	pkRaw, err := base64.StdEncoding.DecodeString(pkRawB64)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}

	pk, err := jwt.ParseECPrivateKeyFromPEM([]byte(pkRaw))
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	return &JWTSigner{
		pk:     pk,
		issuer: issuer,
	}, nil
}

func (s *JWTSigner) Sign(claims model.TokenClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss":       s.issuer,
		"aud":       claims.Audience,
		"sub":       claims.Subject,
		"iat":       claims.CreatedAt,
		"exp":       claims.ExpiresAt,
		"scopes":    claims.Scopes,
		"client_id": claims.ClientID,
	})

	return token.SignedString(s.pk)
}

func (s *JWTSigner) GetPublicKey() string {
	pubKey := s.pk.Public()
	pubKeyDER, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(pubKeyDER)
}

func (s *JWTSigner) GetPublicKeyPEM() string {
	pubKey := s.pk.Public()
	pubKeyDER, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return ""
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyDER,
	}

	return string(pem.EncodeToMemory(pemBlock))
}
