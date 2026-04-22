package fake

import (
	"context"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/application/port"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"go.uber.org/zap"
)

var _ port.OTPSender = (*OTPSender)(nil)

type OTPSender struct {
	logger log.Logger
}

func NewOTPSender(logger log.Logger) port.OTPSender {
	return &OTPSender{
		logger: logger,
	}
}

func (l *OTPSender) Send(ctx context.Context, phone model.Phone, code string) error {
	l.logger.With(
		zap.String("phone", phone.Masked()),
		zap.String("code", code),
	).Info("sending OTP")
	return nil
}

func (l *OTPSender) Resend(ctx context.Context, phone model.Phone, code string) error {
	l.logger.With(
		zap.String("phone", phone.Masked()),
		zap.String("code", code),
	).Info("re-sending OTP")
	return nil
}

func (l *OTPSender) Generate(phone model.Phone, length int) (string, error) {
	l.logger.With(
		zap.String("phone", phone.Masked()),
		zap.Int("length", length),
	).Info("generating OTP")

	const lastDigitsCount = 6

	return string(phone)[len(phone)-lastDigitsCount:], nil
}
