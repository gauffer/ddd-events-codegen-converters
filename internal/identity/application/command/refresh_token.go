package command

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/application/port"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	"fmt"
)

type RefreshTokenCommand struct {
	RefreshToken string
	ClientID     model.OAuthClientID
	ClientSecret *string
	Scopes       model.OAuthScopes

	IP        string
	UserAgent string
}

type RefreshTokenCommandHandler struct {
	clients   repository.OAuthClientsRepository
	sessions  repository.SessionsRepository
	jwtSigner port.JWTSigner
}

func NewRefreshTokenCommandHandler(
	clients repository.OAuthClientsRepository,
	sessions repository.SessionsRepository,
	jwtSigner port.JWTSigner,
) *RefreshTokenCommandHandler {
	return &RefreshTokenCommandHandler{
		clients:   clients,
		sessions:  sessions,
		jwtSigner: jwtSigner,
	}
}

func (h *RefreshTokenCommandHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (*model.TokenPair, error) {
	client, err := h.clients.FindByID(ctx, cmd.ClientID)
	if err != nil {
		return nil, fmt.Errorf("find client by ID: %w", err)
	}

	if client == nil {
		return nil, model.ErrOAuthClientNotFound
	}

	if err := client.ValidateClientSecret(cmd.ClientSecret); err != nil {
		return nil, err
	}

	if !client.AllowsGrantType(model.OAuthGrantTypeRefreshToken) {
		return nil, model.ErrInvalidGrantTypeOAuthClient
	}

	refreshTokenHash := model.HashToken(cmd.RefreshToken)
	session, err := h.sessions.FindByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		return nil, fmt.Errorf("find session by refresh token: %w", err)
	}

	if session == nil {
		return nil, model.ErrSessionIsRevoked
	}

	if session.ClientID != cmd.ClientID {
		return nil, model.ErrSessionClientMismatch
	}

	newRefreshToken, err := model.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := session.RefreshToken(cmd.RefreshToken, newRefreshToken); err != nil {
		if saveErr := h.sessions.Save(ctx, session); saveErr != nil {
			return nil, fmt.Errorf("save session: %w", saveErr)
		}

		return nil, err
	}

	if err := h.sessions.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	tokenClaims, err := client.CreateTokenClaimsForSession(session, cmd.Scopes)
	if err != nil {
		return nil, err
	}

	accessToken, err := h.jwtSigner.Sign(*tokenClaims)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	return &model.TokenPair{
		RefreshToken: newRefreshToken,
		AccessToken:  accessToken,
	}, nil
}
