package ctxUtil

import (
	"context"
	contextKey "jaystar/internal/constant/context"
	"jaystar/internal/utils/auth"
)

func GetUserSessionFromCtx(ctx context.Context) auth.UserSession {
	return GetGenericValueFromCtx[auth.UserSession](ctx, contextKey.UserSession)
}

func GetGenericValueFromCtx[T any](ctx context.Context, key string) T {
	var defVal T
	if ctx == nil {
		return defVal
	}

	value := ctx.Value(key)
	if value == nil {
		return defVal
	}

	m, ok := value.(T)
	if !ok {
		return defVal
	}

	return m
}
