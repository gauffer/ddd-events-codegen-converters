package repository

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=users.go -destination=../mock/users.go -package=mock

type UserRepository interface {
	Create(context.Context, *model.User) error
	Save(context.Context, *model.User) error
	FindByID(context.Context, model.UserID) (*model.User, error)
	FindByPhone(context.Context, model.Phone) (*model.User, error)
}
