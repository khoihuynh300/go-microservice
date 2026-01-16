package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	SKU         string
	Name        string
	Slug        string
	Description string
	CategoryID  uuid.UUID
	Price       float64
	Thumbnail   *string
	Images      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
