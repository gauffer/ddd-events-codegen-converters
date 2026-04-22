// Package contextvalues — хранение значений в context.Context
package contextvalues

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

const (
	RandomBytesCount = 16
)

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestID(ctx context.Context) string {
	res := ctx.Value(requestIDKey{})
	if res == nil {
		return ""
	}

	value, ok := res.(string)
	if !ok {
		return ""
	}

	return value
}

func GenerateRequestID() string {
	b := make([]byte, RandomBytesCount)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
