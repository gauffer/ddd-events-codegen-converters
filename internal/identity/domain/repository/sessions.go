package repository

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=sessions.go -destination=../mock/sessions.go -package=mock

type SessionsRepository interface {
	Create(context.Context, *model.Session) error
	Save(context.Context, *model.Session) error
	FindByID(context.Context, model.SessionID) (*model.Session, error)
	FindByRefreshTokenHash(context.Context, string) (*model.Session, error)
}
