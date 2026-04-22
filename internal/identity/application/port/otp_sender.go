package port

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=otp_sender.go -destination=../mock/otp_sender.go -package=mock

type OTPSender interface {
	Send(ctx context.Context, phone model.Phone, code string) error
	Resend(ctx context.Context, phone model.Phone, code string) error
	Generate(phone model.Phone, length int) (string, error)
}
