package middleware

import (
	"net/http"

	"github.com/google/uuid"
	mdkeys "github.com/khoihuynh300/go-microservice/shared/pkg/metadata"
)

const (
	TraceIDHeader = "X-Trace-Id"
)

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		w.Header().Set(TraceIDHeader, traceID)
		r.Header.Set(mdkeys.TraceIDHeader, traceID)

		next.ServeHTTP(w, r)
	})
}
