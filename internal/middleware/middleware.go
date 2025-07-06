package middleware

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"phpservermanager/internal/config"
)

// Auth provides authentication middleware.
func Auth(cfg config.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Username == "" || cfg.PasswordHash == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok || user != cfg.Username || bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte(pass)) != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
