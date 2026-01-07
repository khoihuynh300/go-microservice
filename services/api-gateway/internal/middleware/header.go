package middleware

import (
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	mdkeys "github.com/khoihuynh300/go-microservice/shared/pkg/const/metadata"
)

func CustomHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case mdkeys.UserIDHeader, mdkeys.TraceIDHeader:
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
