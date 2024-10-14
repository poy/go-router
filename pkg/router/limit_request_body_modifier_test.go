package router_test

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
	injectiontesting "github.com/poy/go-dependency-injection/pkg/injection/testing"
	"github.com/poy/go-router/pkg/router"
)

func TestLimitRequestBody(t *testing.T) {
	t.Parallel()

	router.AddLimitRequestBody(5)
	ctx := injectiontesting.WithTesting(t)
	modifiers := injection.Resolve[injection.Group[router.Modifier]](ctx).Vals()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", strings.NewReader("1234567890"))
	for _, modifier := range modifiers {
		if modifier.Pre == nil {
			continue
		}
		req = modifier.Pre(rec, req)
	}

	data, err := ioutil.ReadAll(req.Body)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if actual, expected := string(data), "12345"; actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
