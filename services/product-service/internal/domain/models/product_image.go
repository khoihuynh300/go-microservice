package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductImage struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	ImageURL  string
	Position  int32
	CreatedAt time.Time
}
