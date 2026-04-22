package identity

import (
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/adapter/inbound/http"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/application/command"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
)

type Module struct {
	httpHandlers *http.IdentityHandlers
	adapters     *Adapters
}

func NewModule(
	logger log.Logger,
	adapters *Adapters,
) *Module {
	handlers := http.NewHandlers(
		logger,
		&http.Commands{
			PhoneInitiate: command.NewPhoneInitiateCommandHandler(
				adapters.otpSessions,
				adapters.users,
				adapters.oauthClients,
				adapters.otpSender,
			),
			PhoneVerify: command.NewPhoneVerifyCommandHandler(
				adapters.otpSessions,
				adapters.authorizationCodes,
				adapters.users,
				logger,
			),
			ExchangeAuthCode: command.NewExchangeAuthCodeCommandHandler(
				adapters.authorizationCodes,
				adapters.sessions,
				adapters.oauthClients,
				adapters.jwtSigner,
				logger,
			),
			RefreshToken: command.NewRefreshTokenCommandHandler(
				adapters.oauthClients,
				adapters.sessions,
				adapters.jwtSigner,
			),
		},
	)

	return &Module{
		httpHandlers: handlers,
		adapters:     adapters,
	}
}

func (m *Module) GetHTTPHandlers() *http.IdentityHandlers {
	return m.httpHandlers
}
