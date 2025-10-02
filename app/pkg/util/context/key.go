package ctxutil

import "context"

type ContextKey[T any] string

func (t ContextKey[T]) Get(ctx context.Context) (res T, ok bool) {
	val := ctx.Value(t)
	if val == nil {
		return res, false
	}
	res, ok = val.(T)
	return
}

func (t ContextKey[T]) Set(ctx context.Context, val T) context.Context {
	return context.WithValue(ctx, t, val)
}
