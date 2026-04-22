package model

import (
	"errors"
	"time"
)

var (
	ErrInvalidPhone               = errors.New("invalid phone number")
	ErrInvalidCodeChallengeMethod = errors.New("invalid code challenge method")
	ErrInvalidSubjectType         = errors.New("invalid subject type")
	ErrInvalidSubject             = errors.New("invalid subject")
	ErrInvalidClientID            = errors.New("invalid client ID")
	ErrInvalidClientSecret        = errors.New("invalid client secret")
	ErrInvalidRedirectURI         = errors.New("invalid redirect URI")
)

var (
	ErrUserBlocked = errors.New("user is blocked")
)

var (
	ErrInvalidNameOAuthClient        = errors.New("invalid name for OAuth client")
	ErrInvalidTypeOAuthClient        = errors.New("invalid type for OAuth client")
	ErrInvalidGrantTypeOAuthClient   = errors.New("invalid grant type for OAuth client")
	ErrSecretRequiredForConfidential = errors.New("secret is required for confidential OAuth client")
	ErrOAuthClientNotFound           = errors.New("oauth client not found")
)

var (
	ErrSessionNotSupportedService      = errors.New("session not supported for service subject type")
	ErrSessionIsRevoked                = errors.New("session is revoked")
	ErrSessionRefreshTokenMismatch     = errors.New("refresh token mismatch")
	ErrSessionClientMismatch           = errors.New("session client mismatch")
	ErrAuthorizationCodeMismatchClient = errors.New("authorization code client mismatch")
	ErrAuthorizationCodeScopesMismatch = errors.New("authorization code scopes mismatch")
)

var (
	ErrAuthorizationCodeExpired     = errors.New("authorization code expired")
	ErrAuthorizationCodeNotFound    = errors.New("authorization code not found")
	ErrAuthorizationCodeScopesEmpty = errors.New("authorization code scopes empty")
)

var (
	ErrOTPExpired               = errors.New("OTP is expired")
	ErrOTPAlreadyVerified       = errors.New("OTP already verified")
	ErrOTPMaxAttemptsReached    = errors.New("maximum OTP attempts reached")
	ErrOTPInvalidCode           = errors.New("invalid OTP code")
	ErrPhoneAuthSessionNotFound = errors.New("phone auth session not found")
	ErrOTPMaxAttemptsRequested  = errors.New("maximum OTP resend attempts reached")
)

type ErrOTPResendTooSoon struct {
	RetryAt time.Time
}

func (e *ErrOTPResendTooSoon) Error() string {
	return "OTP resend too soon"
}
