package router

import (
	"context"
	"net/http"

	"github.com/poy/go-dependency-injection/pkg/injection"
)

func AddContextModifier(f func(context.Context) context.Context) {
	injection.Register[injection.Group[Modifier]](
		func(ctx context.Context) injection.Group[Modifier] {
			return injection.AddToGroup[Modifier](ctx, Modifier{
				Pre: func(rec http.ResponseWriter, req *http.Request) *http.Request {
					return req.WithContext(f(req.Context()))
				},
			})
		})
}
