package router

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/poy/go-dependency-injection/pkg/injection"
	"github.com/poy/go-router/pkg/observability"
)

// AddUserInfoModifier adds a user info modifier to the router. This modifier
// updates the request context with the user info from the request.
// //
// NOTE: This currently is tied directly to the GCP API Gateway. It assumes
// there is a JWT on the X-Apigateway-Api-Userinfo header. See
// https://cloud.google.com/api-gateway/docs/authenticating-users-googleid for
// more information.
func AddUserInfoModifier() {
	injection.Register[injection.Group[Modifier]](
		func(ctx context.Context) injection.Group[Modifier] {
			logger := injection.Resolve[observability.Logger](ctx)

			return injection.AddToGroup[Modifier](ctx, Modifier{
				Pre: func(w http.ResponseWriter, r *http.Request) *http.Request {
					if r.Method == http.MethodOptions {
						return r
					}

					userInfo := r.Header.Get("X-Apigateway-Api-Userinfo")

					data, err := base64.RawURLEncoding.DecodeString(userInfo)
					if err != nil {
						logger.Warnf("invalid user info. base64 decoding failed: %s: %v", userInfo, err)
						return r
					}

					var m map[string]any
					if err := json.Unmarshal(data, &m); err != nil {
						logger.Warnf("invalid user info. JSON unmarshaling failed: %s: %v", userInfo, err)
						return r
					}

					return r.WithContext(WithUserID(r.Context(), fmt.Sprint(m["sub"])))
				},
			})
		})
}

// GetUserInfo returns the user info from the request context.
func GetUserID(ctx context.Context) string {
	user, _ := ctx.Value(userInfoKey{}).(string)
	return user
}

// WithUserID returns a new context with the user info set.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userInfoKey{}, userID)
}

type userInfoKey struct {
}
