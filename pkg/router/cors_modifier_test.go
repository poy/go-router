package router_test

import (
	"net/http/httptest"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
	injectiontesting "github.com/poy/go-dependency-injection/pkg/injection/testing"
	"github.com/poy/go-router/pkg/router"
)

func TestCorsModifier(t *testing.T) {
	t.Parallel()

	router.AddCORSModifier("some-cors")
	ctx := injectiontesting.WithTesting(t)
	modifiers := injection.Resolve[injection.Group[router.Modifier]](ctx).Vals()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	for _, modifier := range modifiers {
		if modifier.Pre == nil {
			continue
		}
		modifier.Pre(rec, req)
	}

	if actual, expected := rec.Header().Get("Access-Control-Allow-Origin"), "some-cors"; actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
