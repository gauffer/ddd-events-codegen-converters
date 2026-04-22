package repository

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=authorization_codes.go -destination=../mock/authorization_codes.go -package=mock

type AuthorizationCodesRepository interface {
	Save(context.Context, *model.AuthorizationCode) error
	Delete(context.Context, model.AuthorizationCodeID) error
	GetByID(context.Context, model.AuthorizationCodeID) (*model.AuthorizationCode, error)
}
