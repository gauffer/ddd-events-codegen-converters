package inmemory

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	"sync"
	"time"
)

var _ repository.OAuthClientsRepository = (*OAuthClientsRepository)(nil)

type OAuthClientsRepository struct {
	clients map[model.OAuthClientID]*model.OAuthClient
	mu      sync.RWMutex
}

func NewOAuthClientsRepository() repository.OAuthClientsRepository {
	clients := make(map[model.OAuthClientID]*model.OAuthClient)

	allowedScopes := []model.OAuthScope{
		"public",
		"challenges:rw",
	}

	clients["1"] = &model.OAuthClient{
		ID:                  "1",
		Name:                "Mobile App",
		Type:                model.OAuthClientTypePublic,
		AllowedGrantTypes:   []model.OAuthGrantType{model.OAuthGrantTypeAuthorizationCode, model.OAuthGrantTypeRefreshToken},
		AllowedScopes:       []model.OAuthScope{"public"},
		AllowedRedirectURIs: []string{},
		Audience:            []string{"one"},
		AccessTokenTTL:      24 * time.Hour,
		RefreshTokenTTL:     30 * 24 * time.Hour,
	}

	clients["admin_zone_client"] = &model.OAuthClient{
		ID:                  "admin_zone_client",
		Name:                "Admin app",
		Type:                model.OAuthClientTypeConfidential,
		AllowedGrantTypes:   []model.OAuthGrantType{model.OAuthGrantTypeAuthorizationCode, model.OAuthGrantTypeRefreshToken},
		AllowedScopes:       allowedScopes,
		AllowedRedirectURIs: []string{},
		Audience:            []string{"two"},
		AccessTokenTTL:      24 * time.Hour,
		RefreshTokenTTL:     30 * 24 * time.Hour,
	}

	return &OAuthClientsRepository{
		clients: clients,
	}
}

func (r *OAuthClientsRepository) FindByID(ctx context.Context, id model.OAuthClientID) (*model.OAuthClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, ok := r.clients[id]
	if !ok {
		return nil, nil
	}

	return client, nil
}
