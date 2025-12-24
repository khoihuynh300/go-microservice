package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/khoihuynh300/go-microservice/api-gateway/internal/config"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/security/jwtvalidator"
	mdkeys "github.com/khoihuynh300/go-microservice/shared/pkg/metadata"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
)

var publicRoutes = []string{
	"/v1/auth/*",
}

func AuthMiddleware(next http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get(AuthorizationHeader)
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing authorization header"})
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid authorization header format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := jwtvalidator.VerifyAccessToken(tokenString, cfg.Secret)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid or expired token"})
			return
		}

		r.Header.Set(mdkeys.UserIDHeader, claims.Subject)

		next.ServeHTTP(w, r)
	})
}

func isPublicRoute(path string) bool {
	for _, route := range publicRoutes {
		if route == path {
			return true
		}
		if strings.HasSuffix(route, "*") {
			prefix := strings.TrimSuffix(route, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}
	return false
}
