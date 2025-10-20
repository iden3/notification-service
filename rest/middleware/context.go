package middleware

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"
)

type fromDIDKey struct{}

// GetDIDFromContext retrieves the fromDID from the request context.
func GetDIDFromContext(ctx context.Context) (w3c.DID, bool) {
	did, ok := ctx.Value(fromDIDKey{}).(w3c.DID)
	return did, ok
}

// WithDIDContext adds the fromDID to the request context.
func WithDIDContext(ctx context.Context, did w3c.DID) context.Context {
	return context.WithValue(ctx, fromDIDKey{}, did)
}
