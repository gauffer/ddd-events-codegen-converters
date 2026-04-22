package repository

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=oauth_clients.go -destination=../mock/oauth_clients.go -package=mock

type OAuthClientsRepository interface {
	FindByID(context.Context, model.OAuthClientID) (*model.OAuthClient, error)
}
