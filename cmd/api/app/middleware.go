package app

import (
	"net/http"
	"strings"

	"github.com/gauffer/ddd-events-codegen-converters/internal/pkg/reqctx"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/zap"
)

type jwsValidator interface {
	ValidateJWS(jws string) (jwt.Token, error)
}

func jwtWithSubject(v jwsValidator, logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}

			token, err := v.ValidateJWS(strings.TrimPrefix(authHeader, "Bearer "))
			if err != nil {
				logger.With(zap.Error(err)).Warn("invalid JWT")
				next.ServeHTTP(w, r)
				return
			}

			if subject := token.Subject(); subject != "" {
				ctx := reqctx.SetUserID(r.Context(), subject)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
