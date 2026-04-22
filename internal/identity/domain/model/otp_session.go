package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OTPSessionID string

const (
	MaxTotalAttempts         = 6
	MaxOTPAttempts           = 3
	OTPResendCooldown        = 60 * time.Second
	OTPDuration              = 5 * time.Minute
	PhoneAuthSessionDuration = 10 * time.Minute
	OTPLength                = 6
)

type OTP struct {
	Code       string
	Attempts   int
	CreatedAt  time.Time
	ExpiresAt  time.Time
	VerifiedAt *time.Time
}

func (o *OTP) Verify(code string) error {
	if o.VerifiedAt != nil {
		return ErrOTPAlreadyVerified
	}

	if time.Now().After(o.ExpiresAt) {
		return ErrOTPExpired
	}

	if o.Attempts >= MaxOTPAttempts {
		return ErrOTPMaxAttemptsReached
	}

	if o.Code != code {
		o.Attempts++
		return ErrOTPInvalidCode
	}

	now := time.Now()
	o.VerifiedAt = &now

	return nil
}

func (o *OTP) IsVerified() bool {
	return o.VerifiedAt != nil
}

func (o *OTP) CanResend() bool {
	return OTPResendCooldown <= time.Since(o.CreatedAt) && o.Attempts < MaxOTPAttempts
}

// OTPSession — сессия аутентификации по OTP
type OTPSession struct {

	ID    OTPSessionID
	Phone Phone

	OTP      *OTP
	Attempts int

	ProofKey ProofKey
	ClientID OAuthClientID
	Scopes   []OAuthScope

	IP        string
	UserAgent string

	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewOTPSession(
	phone Phone,
	proofKey ProofKey,
	clientID OAuthClientID,
	scopes []OAuthScope,
	ip, userAgent string,
) (*OTPSession, error) {
	if !proofKey.Method.IsValid() {
		return nil, ErrInvalidCodeChallengeMethod
	}

	if !phone.IsValid() {
		return nil, ErrInvalidPhone
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("generating phone auth session uuid: %w", err)
	}

	now := time.Now()

	session := &OTPSession{
		ID:    OTPSessionID(id.String()),
		Phone: phone,

		ProofKey: proofKey,
		ClientID: clientID,
		Scopes:   scopes,

		IP:        ip,
		UserAgent: userAgent,

		ExpiresAt: now.Add(PhoneAuthSessionDuration),
		CreatedAt: now,
	}


	return session, nil
}

func (s *OTPSession) GetRetryAt() time.Time {
	if s.OTP == nil {
		return s.CreatedAt.Add(OTPResendCooldown)
	}

	return s.OTP.CreatedAt.Add(OTPResendCooldown)
}

func (s *OTPSession) AssignOTP(code string, ip, userAgent string) error {
	if err := s.validateOTPAssignment(ip, userAgent); err != nil {
		return err
	}

	now := time.Now()
	s.OTP = &OTP{
		Code:      code,
		Attempts:  0,
		CreatedAt: now,
		ExpiresAt: now.Add(OTPDuration),
	}
	s.Attempts++

	return nil
}

// CreateAuthorizationCode — создаёт auth code после верификации
func (s *OTPSession) CreateAuthorizationCode(userID UserID) (*AuthorizationCode, error) {
	if s.OTP == nil || !s.OTP.IsVerified() {
		return nil, ErrOTPInvalidCode
	}

	return NewAuthorizationCode(
		s.ClientID,
		s.Scopes,
		"",
		string(userID),
		SubjectTypeUser,
		&s.ProofKey,
	)
}

// Verify — проверка OTP кода
func (s *OTPSession) Verify(code string, ip, userAgent string) error {
	if s.isExpired() {
		s.recordVerificationFailed(ip, userAgent, "session expired")
		return ErrPhoneAuthSessionNotFound
	}

	if s.OTP == nil {
		s.recordVerificationFailed(ip, userAgent, "OTP not assigned")
		return ErrOTPInvalidCode
	}

	if err := s.OTP.Verify(code); err != nil {
		s.recordVerificationFailed(ip, userAgent, err.Error())
		return err
	}

	
	return nil
}

func (s *OTPSession) validateOTPAssignment(ip, userAgent string) error {
	if s.isExpired() {
		return ErrPhoneAuthSessionNotFound
	}

	if s.OTP == nil {
		return nil
	}

	if s.OTP.IsVerified() {
		return ErrOTPAlreadyVerified
	}

	if s.hasReachedMaxAttempts() {
		s.recordVerificationFailed(ip, userAgent, ErrOTPMaxAttemptsReached.Error())
		return ErrOTPMaxAttemptsRequested
	}

	if !s.canResendOTP() {
		return &ErrOTPResendTooSoon{
			RetryAt: s.GetRetryAt(),
		}
	}

	return nil
}

func (s *OTPSession) isExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *OTPSession) hasReachedMaxAttempts() bool {
	return s.Attempts >= MaxTotalAttempts
}

func (s *OTPSession) canResendOTP() bool {
	if s.OTP == nil {
		return true
	}
	return time.Since(s.OTP.CreatedAt) >= OTPResendCooldown
}

func (s *OTPSession) recordVerificationFailed(ip, userAgent, reason string) {
	
}
