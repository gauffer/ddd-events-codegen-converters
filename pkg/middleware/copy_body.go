package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"go.uber.org/zap"
)

func copyBody(r *http.Request, logger log.Logger) []byte {
	buffer := new(bytes.Buffer)

	_, _ = buffer.ReadFrom(r.Body)
	if err := r.Body.Close(); err != nil {
		logger.With(zap.Error(err)).Warn("error closing request body")
	}

	r.Body = io.NopCloser(buffer)
	return buffer.Bytes()
}
