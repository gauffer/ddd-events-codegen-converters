package http

import "github.com/gauffer/ddd-events-codegen-converters/pkg/api/types"

type (
	OTPResendTooSoonError = types.OTPResendTooSoonError
	ApiError              = types.ApiError
	ApiErrorCode          = types.ApiErrorCode
	PhoneInitiateRequest  = types.PhoneInitiateRequest
	PhoneInitiateResponse = types.PhoneInitiateResponse
	PhoneVerifyRequest    = types.PhoneVerifyRequest
	PhoneVerifyResponse   = types.PhoneVerifyResponse
	TokenResponse         = types.TokenResponse
)

const (
	InternalServerError = types.InternalServerError
)
