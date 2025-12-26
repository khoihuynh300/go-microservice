package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/khoihuynh300/go-microservice/api-gateway/internal/config"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/security/jwtvalidator"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
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
			writeErrorResponse(w, apperr.ErrUnauthenticated)
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			writeErrorResponse(w, apperr.ErrInvalidAuthHeader)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := jwtvalidator.VerifyAccessToken(tokenString, cfg.Secret)
		if err != nil {
			if err == jwtvalidator.ErrTokenExpired {
				writeErrorResponse(w, apperr.ErrTokenExpired)
				return
			} else if err == jwtvalidator.ErrTokenInvalid {
				writeErrorResponse(w, apperr.ErrTokenInvalid)
				return
			}

			writeErrorResponse(w, apperr.ErrInternal)
			return
		}

		r.Header.Set(mdkeys.UserIDHeader, claims.Subject)

		next.ServeHTTP(w, r)
	})
}

func writeErrorResponse(w http.ResponseWriter, err *apperr.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	json.NewEncoder(w).Encode(err)
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
