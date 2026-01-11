package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID             uuid.UUID `json:"id"`
	SKU            string    `json:"sku"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	CategoryID     uuid.UUID `json:"category_id"`
	Price          float64   `json:"price"`
	CompareAtPrice *float64  `json:"compare_at_price,omitempty"`
	Thumbnail      string    `json:"thumbnail"`
	Images         []string  `json:"images"`
	ViewCount      int32     `json:"view_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
