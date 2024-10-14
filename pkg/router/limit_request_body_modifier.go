package router

import (
	"context"
	"net/http"

	"github.com/poy/go-dependency-injection/pkg/injection"
)

// AddLimitRequestBodyModifier is a modifier that limits the request body size.
func AddLimitRequestBody(size int) {
	injection.Register[injection.Group[Modifier]](
		func(ctx context.Context) injection.Group[Modifier] {
			return injection.AddToGroup[Modifier](ctx, Modifier{
				Pre: func(w http.ResponseWriter, r *http.Request) *http.Request {
					r.Body = http.MaxBytesReader(w, r.Body, int64(size))
					return r
				},
			})
		})
}
