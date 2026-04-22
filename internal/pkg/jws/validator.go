package jws

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
)

type Validator interface {
	ValidateJWS(jws string) (jwt.Token, error)
}

type validator struct {
	publicKey *ecdsa.PublicKey
	issuer    string
	audience  string
}

func NewValidator(rawPK, issuer, audience string) (Validator, error) {
	block, _ := pem.Decode([]byte(rawPK))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not ECDSA")
	}

	return &validator{
		publicKey: ecdsaPubKey,
		issuer:    issuer,
		audience:  audience,
	}, nil
}

func (v *validator) ValidateJWS(jws string) (jwt.Token, error) {
	var options []jwt.ParseOption

	options = append(options, jwt.WithVerify(jwa.ES256, v.publicKey))
	options = append(options, jwt.WithValidate(false))

	token, err := jwt.ParseString(jws, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS: %w", err)
	}

	now := time.Now()

	if v.issuer != "" {
		if token.Issuer() != v.issuer {
			return nil, fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, token.Issuer())
		}
	}

	if token.Subject() == "" {
		return nil, errors.New("missing subject (sub) claim")
	}

	if !token.Expiration().IsZero() {
		if token.Expiration().Before(now) {
			return nil, fmt.Errorf("token expired at %s", token.Expiration().Format(time.RFC3339))
		}
	} else {
		return nil, errors.New("missing expiration (exp) claim")
	}

	if !token.IssuedAt().IsZero() {
		if token.IssuedAt().After(now.Add(5 * time.Minute)) {
			return nil, fmt.Errorf("token issued in the future at %s", token.IssuedAt().Format(time.RFC3339))
		}
	}

	if v.audience != "" {
		audiences := token.Audience()
		if len(audiences) == 0 {
			return nil, errors.New("missing audience (aud) claim")
		}
		found := false
		for _, aud := range audiences {
			if aud == v.audience {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid audience: expected %s in %v", v.audience, audiences)
		}
	}

	return token, nil
}
