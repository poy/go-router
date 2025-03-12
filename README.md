# go-router
A simple HTTP router for Go

# Motivation
I wanted a way to create HTTP routes in a codebase using https://github.com/poy/go-dependency-injection

# Example

```go
package main

import (
	"context"
	"net/http"

	"github.com/poy/go-dependency-injection/pkg/injection"
	"github.com/poy/go-router/pkg/router"

	// Register the STDOUT logger.
	_ "github.com/poy/go-router/pkg/observability/cli"
)

func main() {
	ctx := injection.WithInjection(context.Background())
	router := injection.Resolve[router.Router](ctx)
	if err := http.ListenAndServe("localhost:8080", router); err != nil {
		panic(err)
	}
}

func init() {
	injection.Register[injection.Group[router.Route]](func(ctx context.Context) injection.Group[router.Route] {
		return injection.AddToGroup[router.Route](ctx, router.Route{
			Method:      http.MethodGet,
			Path:        "/apis/v1alpha1/status",
			Description: `Status endpoint`,
			Handler:     &handler{},
		})
	})
}

type handler struct {
}

func (h *handler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	// NOP. We just want to return a 200.
}
```
