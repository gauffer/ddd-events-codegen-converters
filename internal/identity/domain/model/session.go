package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	SessionExpiresIn = 24 * 180 * time.Hour
)

type SessionID string

type Session struct {

	ID               SessionID
	Subject          string
	SubjectType      SubjectType
	ClientID         OAuthClientID
	Scopes           OAuthScopes
	RefreshTokenHash string

	IP        string
	UserAgent string

	CreatedAt  time.Time
	LastUsedAt time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

func NewSession(
	authorizationCode AuthorizationCode,
	rawRefreshToken string,
	ip, userAgent string,
) (*Session, error) {
	if err := validateAuthorizationCodeForSession(authorizationCode); err != nil {
		return nil, err
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("generating session uuid: %w", err)
	}

	now := time.Now()

	session := &Session{
		ID:          SessionID(id.String()),
		Subject:     authorizationCode.Subject,
		SubjectType: authorizationCode.SubjectType,
		Scopes:      authorizationCode.Scopes,
		ClientID:    authorizationCode.ClientID,

		IP:               ip,
		UserAgent:        userAgent,
		RefreshTokenHash: HashToken(rawRefreshToken),

		CreatedAt:  now,
		LastUsedAt: now,
		ExpiresAt:  now.Add(SessionExpiresIn),
		RevokedAt:  nil,
	}


	return session, nil
}

// RefreshToken — обновление refresh token с валидацией текущего
func (s *Session) RefreshToken(currentRaw, newRaw string) error {
	now := time.Now()
	if err := s.canRefresh(now, currentRaw); err != nil {
		return err
	}

	s.updateRefreshToken(newRaw, now)
	

	return nil
}

func (s *Session) canRefresh(now time.Time, currentRaw string) error {
	if s.isRevoked() {
		s.recordRefreshFailed(ErrSessionIsRevoked.Error())
		return ErrSessionIsRevoked
	}

	if s.isExpired(now) {
		s.recordRefreshFailed("session expired")
		return ErrSessionIsRevoked
	}

	if !s.isValidRefreshToken(currentRaw) {
		s.recordRefreshFailed(ErrSessionRefreshTokenMismatch.Error())
		return ErrSessionRefreshTokenMismatch
	}

	return nil
}

func (s *Session) isRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) isExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

func (s *Session) isValidRefreshToken(token string) bool {
	return s.RefreshTokenHash == HashToken(token)
}

func (s *Session) updateRefreshToken(newRaw string, now time.Time) {
	s.RefreshTokenHash = HashToken(newRaw)
	s.LastUsedAt = now
	s.ExpiresAt = now.Add(SessionExpiresIn)
}

func (s *Session) recordRefreshFailed(reason string) {
	
}

func validateAuthorizationCodeForSession(code AuthorizationCode) error {
	if code.IsExpired() {
		return ErrAuthorizationCodeExpired
	}

	if !code.SubjectType.IsValid() {
		return ErrInvalidSubjectType
	}

	if code.SubjectType.IsService() {
		return ErrSessionNotSupportedService
	}

	return nil
}
