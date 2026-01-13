package storage

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
)

const (
	MaxImageSize = 5 * 1024 * 1024
)

var (
	AllowedImageTypes = []string{"image/jpeg", "image/jpg", "image/png", "image/webp"}
	AllowedImageExts  = []string{".jpg", ".jpeg", ".png", ".webp"}
)

type ImageValidator struct{}

func NewImageValidator() *ImageValidator {
	return &ImageValidator{}
}

func (v *ImageValidator) Validate(fileHeader *multipart.FileHeader) error {
	if fileHeader.Size > MaxImageSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxImageSize)
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !v.isAllowedExtension(ext) {
		return fmt.Errorf("file extension %s is not allowed. Allowed: %v", ext, AllowedImageExts)
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if !v.isAllowedContentType(contentType) {
		return fmt.Errorf("content type %s is not allowed. Allowed: %v", contentType, AllowedImageTypes)
	}

	return nil
}

func (v *ImageValidator) ValidateFileInfo(filename, contentType string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !v.isAllowedExtension(ext) {
		return fmt.Errorf("file extension %s is not allowed. Allowed: %v", ext, AllowedImageExts)
	}

	if !v.isAllowedContentType(contentType) {
		return fmt.Errorf("content type %s is not allowed. Allowed: %v", contentType, AllowedImageTypes)
	}

	return nil
}

func (v *ImageValidator) isAllowedExtension(ext string) bool {
	for _, allowed := range AllowedImageExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

func (v *ImageValidator) isAllowedContentType(contentType string) bool {
	for _, allowed := range AllowedImageTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}
