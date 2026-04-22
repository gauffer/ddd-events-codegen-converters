package http

import (
	"encoding/json"
	"errors"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/pointer"
	"net/http"

	"go.uber.org/zap"
)

// writeError — ответ RFC-9457 Problem. Для 5xx логируется контекст
func (h *IdentityHandlers) writeError(w http.ResponseWriter, r *http.Request, err error) {
	problem, status := h.mapErrorToProblem(err, r.URL.Path)

	if status >= 500 {
		h.logger.With(
			zap.Error(err),
			zap.String("path", r.URL.Path),
			zap.Int("status", status),
		).Error("domain error occurred")
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem)
}

// mapErrorToProblem — маппинг доменных ошибок в RFC-9457 Problem
func (h *IdentityHandlers) mapErrorToProblem(err error, instance string) (interface{}, int) {
	var errOTPResendTooSoon *model.ErrOTPResendTooSoon
	if errors.As(err, &errOTPResendTooSoon) {
		return OTPResendTooSoonError{
			Code:     "OTPResendTooSoon",
			Detail:   pointer.Ptr("The OTP resend is too soon"),
			Instance: instance,
			Status:   http.StatusTooManyRequests,
			Title:    "OTP Resend Too Soon",
			RetryAt:  errOTPResendTooSoon.RetryAt,
		}, http.StatusTooManyRequests
	}

	switch err {
	case model.ErrOTPInvalidCode:
		return ApiError{
			Code:     ApiErrorCode("OTPInvalidCode"),
			Title:    "Invalid OTP Code",
			Status:   http.StatusBadRequest,
			Detail:   pointer.Ptr("The OTP code is invalid"),
			Instance: instance,
		}, http.StatusBadRequest
	case model.ErrOTPExpired:
		return ApiError{
			Code:     ApiErrorCode("OTPExpired"),
			Title:    "OTP Expired",
			Status:   http.StatusBadRequest,
			Detail:   pointer.Ptr("The OTP code is expired"),
			Instance: instance,
		}, http.StatusBadRequest
	case model.ErrOTPAlreadyVerified:
		return ApiError{
			Code:     ApiErrorCode("OTPAlreadyVerified"),
			Title:    "OTP Already Verified",
			Status:   http.StatusBadRequest,
			Detail:   pointer.Ptr("The OTP code was already verified"),
			Instance: instance,
		}, http.StatusBadRequest
	case model.ErrOTPMaxAttemptsReached:
		return ApiError{
			Code:     ApiErrorCode("OTPMaxAttemptsReached"),
			Title:    "OTP Max Attempts Reached",
			Status:   http.StatusBadRequest,
			Detail:   pointer.Ptr("The maximum OTP attempts reached"),
			Instance: instance,
		}, http.StatusBadRequest
	case model.ErrOTPMaxAttemptsRequested:
		return ApiError{
			Code:     ApiErrorCode("OTPMaxAttemptsRequested"),
			Title:    "OTP Max Attempts Requested",
			Status:   http.StatusBadRequest,
			Detail:   pointer.Ptr("The maximum OTP resend attempts reached"),
			Instance: instance,
		}, http.StatusBadRequest
	case model.ErrOAuthClientNotFound:
		return ApiError{
			Code:     ApiErrorCode("OAuthClientNotFound"),
			Title:    "OAuth Client Not Found",
			Status:   http.StatusNotFound,
			Detail:   pointer.Ptr("The requested OAuth client was not found"),
			Instance: instance,
		}, http.StatusNotFound
	case model.ErrPhoneAuthSessionNotFound:
		return ApiError{
			Code:     ApiErrorCode("PhoneAuthSessionNotFound"),
			Title:    "Phone Auth Session Not Found",
			Status:   http.StatusNotFound,
			Detail:   pointer.Ptr("The phone auth session was not found"),
			Instance: instance,
		}, http.StatusNotFound
	case model.ErrAuthorizationCodeScopesMismatch:
		return ApiError{
			Code:     ApiErrorCode("AuthorizationCodeScopesMismatch"),
			Title:    "Authorization Code Scopes Mismatch",
			Status:   http.StatusBadRequest,
			Detail:   pointer.Ptr("The authorization code scopes mismatch"),
			Instance: instance,
		}, http.StatusBadRequest
	case model.ErrInvalidCodeChallengeMethod:
		return ApiError{
			Code:     ApiErrorCode("InvalidCodeChallengeMethod"),
			Title:    "Invalid Code Challenge Method",
			Status:   http.StatusBadRequest,
			Detail:   pointer.Ptr("The code challenge method is invalid"),
			Instance: instance,
		}, http.StatusBadRequest
	case model.ErrAuthorizationCodeNotFound:
		return ApiError{
			Code:     ApiErrorCode("AuthorizationCodeNotFound"),
			Title:    "Authorization Code Not Found",
			Status:   http.StatusNotFound,
			Detail:   pointer.Ptr("The authorization code was not found"),
			Instance: instance,
		}, http.StatusNotFound
	}

	return ApiError{
		Code:     InternalServerError,
		Title:    "Internal Server Error",
		Status:   http.StatusInternalServerError,
		Detail:   pointer.Ptr("An internal server error occurred"),
		Instance: instance,
	}, http.StatusInternalServerError
}
