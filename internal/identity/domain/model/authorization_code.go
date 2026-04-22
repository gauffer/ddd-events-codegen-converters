package model

import "time"

const (
	AuthorizationCodeExpirationDuration = 10 * time.Minute
)

type AuthorizationCodeID string

// AuthorizationCode — код авторизации OAuth 2.0
type AuthorizationCode struct {
	ID          AuthorizationCodeID
	ClientID    OAuthClientID
	Scopes      OAuthScopes
	Subject     string
	SubjectType SubjectType
	RedirectURI string

	ProofKey *ProofKey

	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewAuthorizationCode(
	clientID OAuthClientID,
	scopes []OAuthScope,
	redirectURI string,
	subject string,
	subjectType SubjectType,
	proofKey *ProofKey,
) (*AuthorizationCode, error) {
	if subject == "" {
		return nil, ErrInvalidSubject
	}

	if !subjectType.IsValid() {
		return nil, ErrInvalidSubjectType
	}

	if len(scopes) == 0 {
		return nil, ErrAuthorizationCodeScopesEmpty
	}

	id, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &AuthorizationCode{
		ID:          AuthorizationCodeID(id),
		ClientID:    clientID,
		Scopes:      scopes,
		RedirectURI: redirectURI,
		Subject:     subject,
		SubjectType: subjectType,
		ProofKey:    proofKey,
		ExpiresAt:   now.Add(AuthorizationCodeExpirationDuration),
		CreatedAt:   now,
	}, nil
}

func (c *AuthorizationCode) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *AuthorizationCode) Verify(source string) bool {
	return c.ProofKey != nil && c.ProofKey.Verify(source)
}

// ValidateClientID — проверка соответствия clientID
func (c *AuthorizationCode) ValidateClientID(clientID OAuthClientID) error {
	if c.ClientID != clientID {
		return ErrAuthorizationCodeMismatchClient
	}
	return nil
}
