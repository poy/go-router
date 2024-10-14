package router

import (
	"context"
	"net/http"

	"github.com/poy/go-dependency-injection/pkg/injection"
)

// AddCorsHeaders adds the CORS headers to the response.
func AddCORSModifier(cors string) {
	injection.Register[injection.Group[Modifier]](
		func(ctx context.Context) injection.Group[Modifier] {
			return injection.AddToGroup[Modifier](ctx, Modifier{
				Pre: func(w http.ResponseWriter, r *http.Request) *http.Request {
					w.Header().Add("Access-Control-Allow-Origin", cors)
					w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
					w.Header().Add("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
					return r
				},
			})
		})
}
