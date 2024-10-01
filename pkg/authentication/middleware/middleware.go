package authmiddleware

import (
	"net/http"

	"github.com/ry461ch/loyalty_system/pkg/authentication"
)

func Authenticate(authenticator *authentication.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqHeaderJWT := r.Header.Get("Authorization")

			userID, err := authenticator.GetUserID(reqHeaderJWT)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			r.Header.Set("X-User-Id", userID.String())
			next.ServeHTTP(w, r)
		})
	}
}
