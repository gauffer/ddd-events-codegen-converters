package reqctx

import (
	"context"
)

type userIDContextKey struct{}

func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}


