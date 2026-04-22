package command

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/application/port"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	"fmt"
)

type PhoneInitiateCommand struct {
	Phone model.Phone

	ProofKey model.ProofKey
	ClientID model.OAuthClientID
	Scopes   []model.OAuthScope

	IP        string
	UserAgent string
}

type PhoneInitiateCommandHandler struct {
	otpSessions repository.OTPSessionsRepository
	users       repository.UserRepository
	clients     repository.OAuthClientsRepository

	otp port.OTPSender
}

func NewPhoneInitiateCommandHandler(
	otpSessions repository.OTPSessionsRepository,
	users repository.UserRepository,
	clients repository.OAuthClientsRepository,
	otp port.OTPSender,
) *PhoneInitiateCommandHandler {
	return &PhoneInitiateCommandHandler{
		otpSessions: otpSessions,
		users:       users,
		clients:     clients,
		otp:         otp,
	}
}

func (h *PhoneInitiateCommandHandler) Handle(ctx context.Context, cmd PhoneInitiateCommand) (*model.OTPSession, error) {
	client, err := h.clients.FindByID(ctx, cmd.ClientID)
	if err != nil {
		return nil, fmt.Errorf("find client by ID: %w", err)
	}

	if client == nil {
		return nil, model.ErrOAuthClientNotFound
	}

	if !client.IsAllowedScopes(cmd.Scopes) {
		return nil, model.ErrAuthorizationCodeScopesMismatch
	}

	user, err := h.users.FindByPhone(ctx, cmd.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by phone: %w", err)
	}

	if user != nil && user.IsBlocked() {
		return nil, model.ErrUserBlocked
	}

	session, err := h.otpSessions.FindActiveByPhone(ctx, cmd.Phone)
	if err != nil {
		return nil, fmt.Errorf("find active session by phone: %w", err)
	}

	if session == nil {
		session, err = model.NewOTPSession(
			cmd.Phone,
			cmd.ProofKey,
			cmd.ClientID,
			cmd.Scopes,
			cmd.IP,
			cmd.UserAgent,
		)
		if err != nil {
			return nil, err
		}
	}

	code, err := h.otp.Generate(cmd.Phone, model.OTPLength)
	if err != nil {
		return nil, fmt.Errorf("generate OTP code: %w", err)
	}

	if err := session.AssignOTP(code, cmd.IP, cmd.UserAgent); err != nil {
		return nil, fmt.Errorf("assign otp session: %w", err)
	}

	if err := h.otpSessions.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("save otp session: %w", err)
	}

	if err := h.otp.Send(ctx, cmd.Phone, code); err != nil {
		return nil, fmt.Errorf("send OTP code: %w", err)
	}

	return session, nil
}
