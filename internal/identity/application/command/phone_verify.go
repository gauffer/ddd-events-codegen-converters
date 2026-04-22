package command

import (
	"context"
	"errors"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	"fmt"

	"go.uber.org/zap"
)

type PhoneVerifyCommand struct {
	SessionID model.OTPSessionID

	Code      string
	IP        string
	UserAgent string
}

type PhoneVerifyCommandHandler struct {
	otpSessions        repository.OTPSessionsRepository
	authorizationCodes repository.AuthorizationCodesRepository
	users              repository.UserRepository
	logger             log.Logger
}

func NewPhoneVerifyCommandHandler(
	otpSessions repository.OTPSessionsRepository,
	authorizationCodes repository.AuthorizationCodesRepository,
	users repository.UserRepository,
	logger log.Logger,
) *PhoneVerifyCommandHandler {
	return &PhoneVerifyCommandHandler{
		otpSessions:        otpSessions,
		authorizationCodes: authorizationCodes,
		users:              users,
		logger:             logger,
	}
}

func (h *PhoneVerifyCommandHandler) Handle(ctx context.Context, cmd PhoneVerifyCommand) (*model.AuthorizationCode, error) {
	session, err := h.otpSessions.FindByID(ctx, cmd.SessionID)
	if err != nil {
		if errors.Is(err, model.ErrPhoneAuthSessionNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("find phone auth session by ID: %w", err)
	}

	if err := session.Verify(cmd.Code, cmd.IP, cmd.UserAgent); err != nil {
		if saveErr := h.otpSessions.Save(ctx, session); saveErr != nil {
			return nil, fmt.Errorf("save phone auth session: %w", saveErr)
		}
		return nil, err
	}

	if err := h.otpSessions.Delete(ctx, session.ID); err != nil {
		h.logger.With(
			zap.String("session_id", string(session.ID)),
			zap.Error(err),
		).Warn("failed to delete phone auth session after verification")
	}

	user, err := h.findOrCreateUser(ctx, session.Phone)
	if err != nil {
		return nil, fmt.Errorf("find or create user: %w", err)
	}

	user.RecordLogin()
	if err := h.users.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	authorizationCode, err := session.CreateAuthorizationCode(user.ID)
	if err != nil {
		return nil, fmt.Errorf("create authorization code: %w", err)
	}

	if err := h.authorizationCodes.Save(ctx, authorizationCode); err != nil {
		return nil, fmt.Errorf("save authorization code: %w", err)
	}

	return authorizationCode, nil
}

func (h *PhoneVerifyCommandHandler) findOrCreateUser(ctx context.Context, phone model.Phone) (*model.User, error) {
	user, err := h.users.FindByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	if user != nil {
		return user, nil
	}

	user, err = model.NewUser(phone)
	if err != nil {
		return nil, err
	}

	if err := h.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
