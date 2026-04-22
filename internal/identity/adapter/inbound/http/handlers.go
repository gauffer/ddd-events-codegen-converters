package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/application/command"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/pointer"
)

type Commands struct {
	PhoneInitiate    *command.PhoneInitiateCommandHandler
	PhoneVerify      *command.PhoneVerifyCommandHandler
	ExchangeAuthCode *command.ExchangeAuthCodeCommandHandler
	RefreshToken     *command.RefreshTokenCommandHandler
}

type IdentityHandlers struct {
	logger   log.Logger
	commands *Commands
}

func NewHandlers(
	logger log.Logger,
	commands *Commands,
) *IdentityHandlers {
	return &IdentityHandlers{
		logger:   logger,
		commands: commands,
	}
}

// PhoneInitiate — инициация авторизации по номеру телефона
func (h *IdentityHandlers) PhoneInitiate(w http.ResponseWriter, r *http.Request) {
	body := PhoneInitiateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, err)
		return
	}

	proofKey, err := model.NewProofKey(body.CodeChallenge, model.S256)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	scopes := make([]model.OAuthScope, len(body.Scopes))
	for i, s := range body.Scopes {
		scopes[i] = model.OAuthScope(s)
	}

	session, err := h.commands.PhoneInitiate.Handle(r.Context(), command.PhoneInitiateCommand{
		Phone:     model.NewPhone(body.Phone),
		ProofKey:  *proofKey,
		ClientID:  model.OAuthClientID(body.ClientId),
		Scopes:    scopes,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PhoneInitiateResponse{
		SessionId:   string(session.ID),
		ExpiredAt:   session.ExpiresAt,
		PhoneMasked: session.Phone.Masked(),
		RetryAt:     session.GetRetryAt(),
	})
}

// ResendPhoneCode — повторная отправка OTP
func (h *IdentityHandlers) ResendPhoneCode(w http.ResponseWriter, r *http.Request) {
	h.writeNotImplemented(w, r)
}

// PhoneVerify — верификация OTP
func (h *IdentityHandlers) PhoneVerify(w http.ResponseWriter, r *http.Request) {
	body := PhoneVerifyRequest{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, err)
		return
	}

	code, err := h.commands.PhoneVerify.Handle(r.Context(), command.PhoneVerifyCommand{
		SessionID: model.OTPSessionID(body.SessionId),
		Code:      body.Otp,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(PhoneVerifyResponse{
		Code:      string(code.ID),
		ExpiredAt: code.ExpiresAt,
	})
}

// Token — обмен OAuth токенов
func (h *IdentityHandlers) Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeError(w, r, err)
		return
	}

	grantType := r.Form.Get("grant_type")

	switch grantType {
	case "authorization_code":
		pair, err := h.commands.ExchangeAuthCode.Handle(r.Context(), command.ExchangeAuthCodeCommand{
			Code:         model.AuthorizationCodeID(r.Form.Get("code")),
			ClientID:     model.OAuthClientID(r.Form.Get("client_id")),
			CodeVerifier: pointer.Ptr(r.Form.Get("code_verifier")),
			IP:           r.RemoteAddr,
			UserAgent:    r.UserAgent(),
		})
		if err != nil {
			h.writeError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  pair.AccessToken,
			RefreshToken: pointer.Ptr(pair.RefreshToken),
			TokenType:    "Bearer",
		})

	case "refresh_token":
		pair, err := h.commands.RefreshToken.Handle(r.Context(), command.RefreshTokenCommand{
			RefreshToken: r.Form.Get("refresh_token"),
			ClientSecret: pointer.Ptr(r.Form.Get("client_secret")),
			ClientID:     model.OAuthClientID(r.Form.Get("client_id")),
			IP:           r.RemoteAddr,
			UserAgent:    r.UserAgent(),
		})
		if err != nil {
			h.writeError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  pair.AccessToken,
			RefreshToken: pointer.Ptr(pair.RefreshToken),
			TokenType:    "Bearer",
		})

	default:
		h.writeError(w, r, fmt.Errorf("unsupported grant_type: %s", grantType))
	}
}

func (h *IdentityHandlers) writeNotImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(ApiError{
		Code:     "NotImplemented",
		Status:   http.StatusNotImplemented,
		Title:    "Not Implemented",
		Detail:   pointer.Ptr("This endpoint is not yet implemented"),
		Instance: r.URL.Path,
	})
}
