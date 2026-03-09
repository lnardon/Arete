package middleware

import (
	"net/http"

	"github.com/lnardon/arete/internal/auth"
)

func JWT(authSvc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.CookieName)
			if err != nil {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			claims, err := authSvc.ValidateClaims(cookie.Value)
			if err != nil {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			ctx := auth.WithUser(r.Context(), auth.AuthUser{
				ID:       claims.UserID,
				Username: claims.Username,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
