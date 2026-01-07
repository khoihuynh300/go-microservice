package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"google.golang.org/grpc/status"
)

func CustomErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	st := status.Convert(err)

	response := apperr.FromGRPCError(err)

	httpStatus := runtime.HTTPStatusFromCode(st.Code())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	json.NewEncoder(w).Encode(response)
}
