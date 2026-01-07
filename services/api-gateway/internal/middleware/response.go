package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var CodeSuccess = "SUCCESS"

type HTTPSuccessResponse struct {
	Code       string      `json:"code"`
	Message    string      `json:"message,omitempty"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalItems int  `json:"total_items"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// custom marshaler to avoid default proto marshaler behavior
type CustomMarshaler struct {
	runtime.JSONPb
}

func (m *CustomMarshaler) Marshal(v interface{}) ([]byte, error) {
	return nil, nil
}

func SuccessResponseModifier(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {

	marshaler := protojson.MarshalOptions{
		UseProtoNames:   false, // camelCase (json_name)
		EmitUnpopulated: true,
	}

	jsonBytes, err := marshaler.Marshal(resp)
	if err != nil {
		return err
	}

	var data any
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return err
	}

	response := HTTPSuccessResponse{
		Code: CodeSuccess,
		Data: data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	return nil
}
