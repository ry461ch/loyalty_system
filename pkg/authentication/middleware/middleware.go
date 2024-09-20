package authmiddleware

import (
	"net/http"

	"github.com/ry461ch/loyalty_system/pkg/authentication"
)

func Authenticate(authenticator *authentication.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqHeaderJWT := r.Header.Get("Authorization")

			userId, err := authenticator.GetUserId(reqHeaderJWT)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			r.Header.Set("X-User-Id", userId.String())
			next.ServeHTTP(w, r)
		})
	}
}
