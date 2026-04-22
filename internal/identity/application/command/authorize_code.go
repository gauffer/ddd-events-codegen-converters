package command

import (
	"context"
	"errors"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/application/port"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	"fmt"

	"go.uber.org/zap"
)

type ExchangeAuthCodeCommand struct {
	Code         model.AuthorizationCodeID
	ClientID     model.OAuthClientID
	CodeVerifier *string

	IP        string
	UserAgent string
}

type ExchangeAuthCodeCommandHandler struct {
	authorizationCodes repository.AuthorizationCodesRepository
	sessions           repository.SessionsRepository
	clients            repository.OAuthClientsRepository
	jwtSigner          port.JWTSigner
	logger             log.Logger
}

func NewExchangeAuthCodeCommandHandler(
	authorizationCodes repository.AuthorizationCodesRepository,
	sessions repository.SessionsRepository,
	clients repository.OAuthClientsRepository,
	jwtSigner port.JWTSigner,
	logger log.Logger,
) *ExchangeAuthCodeCommandHandler {
	return &ExchangeAuthCodeCommandHandler{
		authorizationCodes: authorizationCodes,
		sessions:           sessions,
		clients:            clients,
		jwtSigner:          jwtSigner,
		logger:             logger,
	}
}

func (h *ExchangeAuthCodeCommandHandler) Handle(ctx context.Context, cmd ExchangeAuthCodeCommand) (*model.TokenPair, error) {
	code, err := h.authorizationCodes.GetByID(ctx, cmd.Code)
	if err != nil {
		if errors.Is(err, model.ErrAuthorizationCodeNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("get authorization code by ID: %w", err)
	}

	client, err := h.clients.FindByID(ctx, code.ClientID)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, model.ErrOAuthClientNotFound
	}

	tokenClaims, err := client.ExchangeCode(*code, cmd.CodeVerifier, code.RedirectURI, cmd.IP, cmd.UserAgent)
	if err != nil {
		return nil, err
	}

	refreshToken, err := model.GenerateToken()
	if err != nil {
		return nil, err
	}

	accessToken, err := h.jwtSigner.Sign(*tokenClaims)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	session, err := model.NewSession(
		*code,
		refreshToken,
		cmd.IP,
		cmd.UserAgent,
	)
	if err != nil {
		return nil, err
	}

	if err := h.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	if err := h.authorizationCodes.Delete(ctx, code.ID); err != nil {
		h.logger.With(
			zap.String("code_id", string(code.ID)),
			zap.Error(err),
		).Warn("failed to delete authorization code after exchange")
	}

	return &model.TokenPair{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}
