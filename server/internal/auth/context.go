package auth

import "context"

type contextKey struct{}

type AuthUser struct {
	ID       string
	Username string
}

func WithUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

func UserFromContext(ctx context.Context) (AuthUser, bool) {
	user, ok := ctx.Value(contextKey{}).(AuthUser)
	return user, ok
}
