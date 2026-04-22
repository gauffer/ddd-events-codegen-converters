package types

import "time"

const (
	InternalServerError = ApiErrorCode("InternalServerError")
)

type ApiErrorCode string

type ApiError struct {
	Code     ApiErrorCode `json:"code"`
	Status   int          `json:"status"`
	Title    string       `json:"title"`
	Detail   *string      `json:"detail,omitempty"`
	Instance string       `json:"instance"`
}

type OTPResendTooSoonError struct {
	Code     string    `json:"code"`
	Status   int       `json:"status"`
	Title    string    `json:"title"`
	Detail   *string   `json:"detail,omitempty"`
	Instance string    `json:"instance"`
	RetryAt  time.Time `json:"retry_at"`
}

type PhoneInitiateRequest struct {
	Phone               string   `json:"phone"`
	Scopes              []string `json:"scopes"`
	ClientId            string   `json:"client_id"`
	CodeChallenge       string   `json:"code_challenge"`
	CodeChallengeMethod string   `json:"code_challenge_method"`
}

type PhoneInitiateResponse struct {
	SessionId   string    `json:"session_id"`
	ExpiredAt   time.Time `json:"expired_at"`
	PhoneMasked string    `json:"phone_masked"`
	RetryAt     time.Time `json:"retry_at"`
}

type PhoneVerifyRequest struct {
	SessionId string `json:"session_id"`
	Otp       string `json:"otp"`
}

type PhoneVerifyResponse struct {
	Code      string    `json:"code"`
	ExpiredAt time.Time `json:"expired_at"`
}

type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken *string `json:"refresh_token,omitempty"`
	TokenType    string  `json:"token_type"`
}
