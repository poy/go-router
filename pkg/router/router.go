package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/poy/go-dependency-injection/pkg/injection"
	"github.com/poy/go-router/pkg/observability"
)

func init() {
	injection.Register[Router](newRouter)
}

// Router is an HTTP router.
type Router http.Handler

// Route is a route that will be registered with the HTTP router. To use the
// Router, you must register a Group[Route] with the injection system.
type Route struct {
	Method      string
	Path        string
	Handler     http.Handler
	Description string

	// TODO(poy): It would be nice if the router could enforce this instead of
	// just adding it to the OpenAPI V3 spec.
	RequiredHeaders map[string]string
	ResponseSchema  any
}

// Modifier is used to modify each Request/Response into the Router.
type Modifier struct {
	// Pre is invoked before the main ServeHTTP function if non-nil.
	Pre func(http.ResponseWriter, *http.Request) *http.Request
}

func newRouter(ctx context.Context) Router {
	routes := injection.Resolve[injection.Group[Route]](ctx).Vals()
	logger := injection.Resolve[observability.Logger](ctx)
	allowedMethods := make(map[string][]string)
	modify := setupModifiers(ctx)

	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Warnf("not found: %s:%s", r.Method, r.URL.String())
		WriteError(w, http.StatusNotFound, fmt.Errorf("path %s not found", r.URL.Path))
	})
	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s not allowed", r.Method))
	})

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path > routes[j].Path
	})

	for _, r := range routes {
		// Avoid issues with closure.
		r := r

		logger.Infof("Registering route: %s %s", r.Method, r.Path)

		handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req = modify(w, req)
			req = req.WithContext(withPathVars(req.Context(), mux.Vars(req)))
			r.Handler.ServeHTTP(w, req)
		})
		router.Handle(r.Path, handler).Methods(r.Method)
		allowedMethods[r.Path] = append(allowedMethods[r.Path], r.Method)
	}

	for path, methods := range allowedMethods {
		methodsStr := strings.Join(methods, ",")
		router.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Allow", methodsStr)
			modify(w, req)
		})).Methods(http.MethodOptions)
	}
	return router
}

func setupModifiers(ctx context.Context) func(http.ResponseWriter, *http.Request) *http.Request {
	g, _ := injection.TryResolve[injection.Group[Modifier]](ctx)
	ms := g.Vals()

	return func(w http.ResponseWriter, r *http.Request) *http.Request {
		for _, m := range ms {
			if m.Pre == nil {
				continue
			}
			r = m.Pre(w, r)
		}
		return r
	}
}

// WriteError writes an error to the response as JSON.
func WriteError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	}); err != nil {
		log.Panicf("failed to marshal error: %v", err)
	}
}

// ReadRequest reads a request from the body and unmarshals it into the given
// Type.
func ReadRequest[TReq any](r *http.Request) (TReq, error) {
	defer r.Body.Close()
	var req TReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, err
	}
	return req, nil
}

// WriteResponse writes a response to the ResponseWriter. If writing fails, it
// will send a 500, however this is likely to not arrive to the user given
// writing the initial response failed. Therefore, this function does not
// return an error, as nothing actinally useful can be done with it.
func WriteResponse[TReq any](w http.ResponseWriter, data TReq) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to write response: %v", err))
		return
	}
}

type pathVarKey struct{}

// PathVars returns the path variables from the request context.
func PathVarsFromContext(ctx context.Context) map[string]string {
	if vars, ok := ctx.Value(pathVarKey{}).(map[string]string); ok {
		return vars
	}
	panic("path vars not found in context. This function can only be used from a request context from the router.")
}

func withPathVars(ctx context.Context, vars map[string]string) context.Context {
	return context.WithValue(ctx, pathVarKey{}, vars)
}
