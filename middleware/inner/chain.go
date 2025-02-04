// Goroutine middlewares
package inner

import (
	"context"
	"errors"
)

type InternalMiddlewareFn[T any] func(ctx context.Context) (T, error)

type InternalMiddleware[T any] func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T]

var (
	ErrUnableToCast = errors.New("interface cast error")
)

func InternalMiddlewareChain[T any](mws ...InternalMiddleware[T]) InternalMiddleware[T] {
	return func(next InternalMiddlewareFn[T]) InternalMiddlewareFn[T] {
		fn := next
		for mw := len(mws) - 1; mw >= 0; mw-- {
			fn = mws[mw](fn)
		}

		return fn
	}
}
