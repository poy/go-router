package router_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
	injectiontesting "github.com/poy/go-dependency-injection/pkg/injection/testing"
	"github.com/poy/go-router/pkg/router"
)

func TestAddUserInfoModifier(t *testing.T) {
	t.Parallel()

	router.AddUserInfoModifier()
	ctx := injectiontesting.WithTesting(t)
	modifiers := injection.Resolve[injection.Group[router.Modifier]](ctx).Vals()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	token := base64.RawStdEncoding.EncodeToString([]byte(`{"sub": "some-user"}`))
	req.Header.Set("X-Apigateway-Api-Userinfo", token)

	for _, modifier := range modifiers {
		if modifier.Pre == nil {
			continue
		}
		req = modifier.Pre(rec, req)
	}

	if expected, actual := "some-user", router.GetUserID(req.Context()); expected != actual {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func TestAddUserInfoModifier_options(t *testing.T) {
	t.Parallel()

	router.AddUserInfoModifier()
	ctx := injectiontesting.WithTesting(t)
	modifiers := injection.Resolve[injection.Group[router.Modifier]](ctx).Vals()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)

	token := base64.RawStdEncoding.EncodeToString([]byte(`{"sub": "some-user"}`))
	req.Header.Set("X-Apigateway-Api-Userinfo", token)

	for _, modifier := range modifiers {
		if modifier.Pre == nil {
			continue
		}
		req = modifier.Pre(rec, req)
	}

	if expected, actual := "", router.GetUserID(req.Context()); expected != actual {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
