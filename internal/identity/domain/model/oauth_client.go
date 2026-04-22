package model

import (
	"time"
)

type TokenClaims struct {
	Subject     string
	Audience    []string
	ClientID    OAuthClientID
	SubjectType SubjectType
	Scopes      OAuthScopes
	CreatedAt   int64
	ExpiresAt   int64
}

// OAuthClient — клиент OAuth 2.0
type OAuthClient struct {

	ID         OAuthClientID
	SecretHash *string
	Name       string
	Type       OAuthClientType

	AllowedGrantTypes   []OAuthGrantType
	AllowedScopes       OAuthScopes
	AllowedRedirectURIs []string

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	Audience []string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *OAuthClient) IsConfidential() bool {
	return c.Type.IsConfidential()
}

func (c *OAuthClient) IsRequiredPKCE() bool {
	return !c.Type.IsConfidential()
}

func (c *OAuthClient) IsAllowedScopes(requestedScopes OAuthScopes) bool {
	if len(requestedScopes) == 0 {
		return true
	}

	return c.AllowedScopes.ContainsAll(requestedScopes)
}

func (c *OAuthClient) Authorize(
	code AuthorizationCode,
	redirectURI string,
	ip, userAgent string,
) (*TokenClaims, error) {
	if code.ClientID != c.ID {
		return nil, ErrInvalidClientID
	}
	if !code.SubjectType.IsValid() {
		return nil, ErrInvalidSubjectType
	}

	if !c.AllowedScopes.ContainsAll(code.Scopes) {
		return nil, ErrAuthorizationCodeScopesMismatch
	}

	if redirectURI != "" && !c.isRedirectURIAllowed(redirectURI) {
		return nil, ErrInvalidRedirectURI
	}

	if !c.AllowsGrantType(OAuthGrantTypeAuthorizationCode) {
		return nil, ErrInvalidGrantTypeOAuthClient
	}

	now := time.Now()

	return &TokenClaims{
		Subject:     code.Subject,
		SubjectType: code.SubjectType,
		Audience:    c.Audience,
		ClientID:    c.ID,
		Scopes:      code.Scopes,
		CreatedAt:   now.Unix(),
		ExpiresAt:   now.Add(c.AccessTokenTTL).Unix(),
	}, nil
}

func (c *OAuthClient) isRedirectURIAllowed(redirectURI string) bool {
	for _, allowedURI := range c.AllowedRedirectURIs {
		if allowedURI == redirectURI {
			return true
		}
	}
	return false
}

func (c *OAuthClient) AllowsGrantType(grantType OAuthGrantType) bool {
	for _, gt := range c.AllowedGrantTypes {
		if gt == grantType {
			return true
		}
	}
	return false
}

// ValidateClientSecret — валидация secret для confidential клиентов
func (c *OAuthClient) ValidateClientSecret(secretRaw *string) error {
	if !c.IsConfidential() {
		return nil
	}

	if secretRaw == nil {
		return ErrSecretRequiredForConfidential
	}

	if c.SecretHash == nil {
		return ErrInvalidClientSecret
	}

	if *c.SecretHash != HashToken(*secretRaw) {
		return ErrInvalidClientSecret
	}

	return nil
}

// ExchangeCode — обмен authorization code на токены (PKCE)
func (c *OAuthClient) ExchangeCode(
	code AuthorizationCode,
	codeVerifier *string,
	redirectURI string,
	ip, userAgent string,
) (*TokenClaims, error) {
	if code.IsExpired() {
		return nil, ErrAuthorizationCodeExpired
	}

	if code.ClientID != c.ID {
		return nil, ErrAuthorizationCodeMismatchClient
	}

	if c.IsRequiredPKCE() {
		if codeVerifier == nil {
			return nil, ErrInvalidCodeChallengeMethod
		}
		if code.ProofKey == nil || !code.ProofKey.Verify(*codeVerifier) {
			return nil, ErrInvalidCodeChallengeMethod
		}
	}

	return c.Authorize(code, redirectURI, ip, userAgent)
}

// CreateTokenClaimsForSession — claims для токена из сессии
func (c *OAuthClient) CreateTokenClaimsForSession(
	session *Session,
	requestedScopes OAuthScopes,
) (*TokenClaims, error) {
	finalScopes := requestedScopes
	if len(requestedScopes) == 0 {
		finalScopes = session.Scopes
	} else if !session.Scopes.ContainsAll(requestedScopes) {
		return nil, ErrAuthorizationCodeScopesMismatch
	}

	now := session.LastUsedAt

	return &TokenClaims{
		Subject:     session.Subject,
		SubjectType: session.SubjectType,
		Audience:    c.Audience,
		ClientID:    session.ClientID,
		Scopes:      finalScopes,
		CreatedAt:   now.Unix(),
		ExpiresAt:   now.Add(c.AccessTokenTTL).Unix(),
	}, nil
}
