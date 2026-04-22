package repository

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/domain/model"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=profiles.go -destination=../mock/profiles_repository_mock.go -package=mock

type ProfilesRepository interface {
	FindByUserID(ctx context.Context, userID string) (*model.Profile, error)
	Save(ctx context.Context, profile *model.Profile) error
}
