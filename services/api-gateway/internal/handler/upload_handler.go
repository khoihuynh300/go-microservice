package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/config"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/utils"
	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	"github.com/khoihuynh300/go-microservice/shared/pkg/storage"
)

type UploadHandler struct {
	imageStorage storage.Storage
	validator    *storage.ImageValidator
}

func NewUploadHandler(
	imageStorage storage.Storage,
) *UploadHandler {
	return &UploadHandler{
		imageStorage: imageStorage,
		validator:    storage.NewImageValidator(),
	}
}

func (h *UploadHandler) GetAvatarPresignedURL(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(contextkeys.UserIDKey).(string)
	userFolder := fmt.Sprintf("%s/%s", utils.AvatarFolder, userId)
	h.getPresignedURL(w, r, userFolder)
}

func (h *UploadHandler) GetProductImagePresignedURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["product_id"]
	productFolder := fmt.Sprintf("%s/%s", utils.ProductImageFolder, productID)

	h.getPresignedURL(w, r, productFolder)
}

func (h *UploadHandler) GetCategoryImagePresignedURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["category_id"]
	categoryFolder := fmt.Sprintf("%s/%s", utils.CategoryImageFolder, categoryID)

	h.getPresignedURL(w, r, categoryFolder)
}

func (h *UploadHandler) getPresignedURL(w http.ResponseWriter, r *http.Request, folder string) {
	var req struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validator.ValidateFileInfo(req.Filename, req.ContentType); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	presignedOutput, err := h.imageStorage.GetPresignedUploadURL(r.Context(), &storage.PresignedUploadInput{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Folder:      folder,
		ExpiresIn:   config.GetPresignedURLExpiry(),
	})
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to generate presigned URL")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"upload_url": presignedOutput.UploadURL,
		"final_url":  presignedOutput.FinalURL,
		"key":        presignedOutput.Key,
		"expires_in": int64(presignedOutput.ExpiresIn.Seconds()),
	})
}

func (h *UploadHandler) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *UploadHandler) respondError(w http.ResponseWriter, statusCode int, message string) {
	h.respondJSON(w, statusCode, map[string]string{"error": message})
}
