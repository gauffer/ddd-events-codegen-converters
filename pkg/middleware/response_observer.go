package middleware

import (
	"net/http"

	"github.com/pkg/errors"
)

type responseObserver struct {
	http.ResponseWriter

	status       int
	written      int64
	wroteHeader  bool
	copyBody     bool
	responseBody []byte
}

func (o *responseObserver) Write(p []byte) (int, error) {
	if !o.wroteHeader {
		o.WriteHeader(http.StatusOK)
	}
	if o.copyBody {
		o.responseBody = append(o.responseBody, p...)
	}

	n, err := o.ResponseWriter.Write(p)
	o.written += int64(n)
	return n, errors.WithStack(err)
}

func (o *responseObserver) WriteHeader(code int) {
	o.ResponseWriter.WriteHeader(code)
	if o.wroteHeader {
		return
	}
	o.wroteHeader = true
	o.status = code
}

func (o *responseObserver) StatusCode() int {
	return o.status
}
