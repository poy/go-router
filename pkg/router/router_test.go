package router_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/poy/go-dependency-injection/pkg/injection"
	injectiontesting "github.com/poy/go-dependency-injection/pkg/injection/testing"
	"github.com/poy/go-router/pkg/router"
)

func init() {
	injection.Register[injection.Group[router.Route]](func(ctx context.Context) injection.Group[router.Route] {
		return injection.AddToGroup[router.Route](ctx, router.Route{
			Path:   "/foo",
			Method: http.MethodGet,
			Handler: assertPreHeaderExists(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		})
	})
	injection.Register[injection.Group[router.Route]](func(ctx context.Context) injection.Group[router.Route] {
		return injection.AddToGroup[router.Route](ctx, router.Route{
			Path:   "/bar",
			Method: http.MethodDelete,
			Handler: assertPreHeaderExists(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		})
	})
	injection.Register[injection.Group[router.Route]](func(ctx context.Context) injection.Group[router.Route] {
		return injection.AddToGroup[router.Route](ctx, router.Route{
			Path:   "/baz/{id}",
			Method: http.MethodGet,
			Handler: assertPreHeaderExists(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := router.PathVarsFromContext(r.Context())
				if vars["id"] != "123" {
					w.WriteHeader(http.StatusBadRequest)
				}
				w.WriteHeader(http.StatusOK)
			})),
		})
	})
	injection.Register[injection.Group[router.Route]](func(ctx context.Context) injection.Group[router.Route] {
		return injection.AddToGroup[router.Route](ctx, router.Route{
			Path:   "/baz",
			Method: http.MethodGet,
			Handler: assertPreHeaderExists(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusIMUsed)
			})),
		})
	})

	injection.Register[injection.Group[router.Modifier]](
		func(ctx context.Context) injection.Group[router.Modifier] {
			return injection.AddToGroup[router.Modifier](ctx, router.Modifier{
				Pre: func(w http.ResponseWriter, r *http.Request) *http.Request {
					w.Header().Set("xyz", "*")
					return r
				},
			})
		})
}

func TestRouter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		req    *http.Request
		assert func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "root path",
			req:  buildRequest(http.MethodGet, "/"),
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				expectedStatusCode(t, rec, http.StatusNotFound)
				expectedContentType(t, rec, "application/json")
			},
		},
		{
			name: "POST /foo",
			req:  buildRequest(http.MethodPost, "/foo"),
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				expectedStatusCode(t, rec, http.StatusMethodNotAllowed)
				expectedContentType(t, rec, "application/json")
			},
		},
		{
			name: "GET /foo",
			req:  buildRequest(http.MethodGet, "/foo"),
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				expectedStatusCode(t, rec, http.StatusOK)
				expectedContentType(t, rec, "")
			},
		},
		{
			name: "GET /bar",
			req:  buildRequest(http.MethodGet, "/bar"),
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				expectedStatusCode(t, rec, http.StatusMethodNotAllowed)
				// expectedContentType(t, rec, "application/json")
			},
		},
		{
			name: "DELETE /bar",
			req:  buildRequest(http.MethodDelete, "/bar"),
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				expectedStatusCode(t, rec, http.StatusOK)
				expectedContentType(t, rec, "")
			},
		},
		{
			name: "GET /baz/123",
			req:  buildRequest(http.MethodGet, "/baz/123"),
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				expectedStatusCode(t, rec, http.StatusOK)
				expectedContentType(t, rec, "")
			},
		},
		{
			name: "GET /baz",
			req:  buildRequest(http.MethodGet, "/baz"),
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				expectedStatusCode(t, rec, http.StatusIMUsed)
				expectedContentType(t, rec, "")
			},
		},
		{
			name: "OPTIONS /baz",
			req:  buildRequest(http.MethodOptions, "/baz"),
			assert: func(t *testing.T, rec *httptest.ResponseRecorder) {
				expectedStatusCode(t, rec, http.StatusOK)
				if actual, expected := rec.Header().Get("Allow"), http.MethodGet; actual != expected {
					t.Fatalf("expected %s, got %s", expected, actual)
				}
			},
		},
	}

	for _, tc := range testCases {
		// Avoid issues with closure.
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := injectiontesting.WithTesting(t)
			r := injection.Resolve[router.Router](ctx)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, tc.req)
			tc.assert(t, recorder)
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	router.WriteError(recorder, http.StatusNotFound, errors.New("not found"))
	expectedStatusCode(t, recorder, http.StatusNotFound)
	expectedContentType(t, recorder, "application/json")

	var m map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if actual, expected := m, map[string]string{"error": "not found"}; !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected body to be %q, but was %q", expected, actual)
	}
}

func TestReadRequest(t *testing.T) {
	t.Parallel()
	req := &http.Request{
		Body: &checkCloser{r: strings.NewReader(`{"foo": 1}`)},
	}
	m, err := router.ReadRequest[map[string]int](req)
	if err != nil {
		t.Fatal(err)
	}

	if actual, expected := m, map[string]int{"foo": 1}; !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected body to be %q, but was %q", expected, actual)
	}

	if !req.Body.(*checkCloser).closed {
		t.Error("expected body to be closed")
	}
}

func TestReadRequest_invalidBody(t *testing.T) {
	t.Parallel()
	req := &http.Request{
		Body: ioutil.NopCloser(strings.NewReader("invalid")),
	}
	if _, err := router.ReadRequest[int](req); err == nil {
		t.Fatal("expected an error")
	}
}

func TestWriteRequest(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	router.WriteResponse(recorder, map[string]int{"foo": 1})

	var actual map[string]int
	if err := json.NewDecoder(recorder.Body).Decode(&actual); err != nil {
		t.Fatal(err)
	}

	if expected := map[string]int{"foo": 1}; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", expected, actual)
	}

	if expected, actual := "application/json", recorder.Header().Get("Content-Type"); expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestWriteRequest_EncodingFails(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	invalidPayload := func() {}
	router.WriteResponse(recorder, invalidPayload)

	expectedStatusCode(t, recorder, http.StatusInternalServerError)
}

func buildRequest(method, path string) *http.Request {
	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}

	return &http.Request{
		Method: method,
		URL:    u,
	}
}

func expectedStatusCode(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()

	if actual := rec.Code; actual != expected {
		t.Fatalf("expected status code %d, got %d", expected, actual)
	}
}

func expectedContentType(t *testing.T, rec *httptest.ResponseRecorder, mimeType string) {
	t.Helper()

	if actual, expected := rec.Header().Get("Content-Type"), mimeType; actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func assertPreHeaderExists(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if actual, expected := w.Header().Get("xyz"), "*"; expected != actual {
			panic(fmt.Sprintf("expected %s, got %s", expected, actual))
		}
		handler.ServeHTTP(w, r)
	})
}

type checkCloser struct {
	closed bool
	r      io.Reader
}

func (c *checkCloser) Close() error {
	c.closed = true
	return nil
}

func (c *checkCloser) Read(data []byte) (int, error) {
	return c.r.Read(data)
}
