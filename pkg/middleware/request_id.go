package middleware

import (
	"context"
	"net/http"

	"github.com/gauffer/ddd-events-codegen-converters/pkg/contextvalues"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"go.uber.org/zap"
)

const (
	RequestIDHeaderKey = "X-Request-Id"
)

func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return contextvalues.WithRequestID(ctx, id)
}

func GetRequestIDFromContext(ctx context.Context) string {
	return contextvalues.RequestID(ctx)
}

// RequestID — добавляет request id в контекст запроса
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(RequestIDHeaderKey)
			if requestID == "" {
				requestID = contextvalues.GenerateRequestID()
				r.Header.Set(RequestIDHeaderKey, requestID)
			}
			ctx := log.NewContext(ContextWithRequestID(r.Context(), requestID), zap.String("request_id", requestID))
			w.Header().Set(RequestIDHeaderKey, requestID)
			r.Header.Set(RequestIDHeaderKey, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
