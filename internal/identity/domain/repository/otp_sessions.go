package repository

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=otp_sessions.go -destination=../mock/otp_sessions.go -package=mock

type OTPSessionsRepository interface {
	Save(context.Context, *model.OTPSession) error
	FindByID(context.Context, model.OTPSessionID) (*model.OTPSession, error)
	FindActiveByPhone(context.Context, model.Phone) (*model.OTPSession, error)
	Delete(context.Context, model.OTPSessionID) error
}
