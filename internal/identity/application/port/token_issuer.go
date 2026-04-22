package port

import "github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"

//go:generate go run go.uber.org/mock/mockgen@latest -source=token_issuer.go -destination=../mock/token_issuer.go -package=mock

type JWTSigner interface {
	Sign(model.TokenClaims) (string, error)
}
