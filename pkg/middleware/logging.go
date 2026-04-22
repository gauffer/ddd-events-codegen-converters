package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"go.uber.org/zap"
)

func Logger(logger log.Logger) func(http.Handler) http.Handler {
	return filterLogger(logger, nil)
}

type ResponseObserver interface {
	StatusCode() int
}

func filterLogger(
	logger log.Logger,
	filter func(o ResponseObserver, r *http.Request) bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			o := &responseObserver{ResponseWriter: w}
			requestID := GetRequestIDFromContext(r.Context())

			start := time.Now()
			next.ServeHTTP(o, r)
			if filter == nil || filter(o, r) {
				logger.WithContext(r.Context()).With([]zap.Field{
					zap.String("request", fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, r.Proto)),
					zap.String("request_method", r.Method),
					zap.String("request_uri", r.RequestURI),
					zap.String("request_proto", r.Proto),
					zap.Int("request_duration_ms", int(time.Since(start).Milliseconds())),
					zap.Int("status", o.status),
					zap.Int64("content_length", o.written),
					zap.String("real_ip", r.Header.Get("X-Real-IP")),
					zap.String("proxy_add_x_forwarded_for", r.Header.Get("X-Forwarded-For")),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("http_referrer", r.Referer()),
					zap.String("http_user_agent", r.UserAgent()),
					zap.String("request_id", requestID),
				}...).Info("http request handled")
			}
		})
	}
}
