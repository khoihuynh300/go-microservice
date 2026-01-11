package dto

type CreateCategoryDTO struct {
	ParentID    *string
	Name        string
	Slug        string
	Description string
}

type UpdateCategoryDTO struct {
	ID          string
	ParentID    *string
	Name        *string
	Slug        *string
	Description *string
}
