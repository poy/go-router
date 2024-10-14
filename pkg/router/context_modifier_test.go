package router_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
	injectiontesting "github.com/poy/go-dependency-injection/pkg/injection/testing"
	"github.com/poy/go-router/pkg/router"
)

func TestAddContextModifier(t *testing.T) {
	t.Parallel()

	router.AddContextModifier(func(ctx context.Context) context.Context {
		return context.WithValue(ctx, "foo", "bar")
	})
	ctx := injectiontesting.WithTesting(t)
	modifiers := injection.Resolve[injection.Group[router.Modifier]](ctx).Vals()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	for _, modifier := range modifiers {
		if modifier.Pre == nil {
			continue
		}
		req = modifier.Pre(rec, req)
	}

	if expected, actual := "bar", req.Context().Value("foo").(string); expected != actual {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
