package authctx

import "context"

type ctxKey int

const (
	userIDKey ctxKey = iota
	roleKey
)

// WithUserID returns a copy of ctx carrying the authenticated user id.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithRole returns a copy of ctx carrying the authenticated role string.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

// UserID returns the user id from ctx, if present.
func UserID(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(userIDKey).(int64)
	return v, ok
}

// Role returns the role from ctx, if present.
func Role(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(roleKey).(string)
	return v, ok
}
