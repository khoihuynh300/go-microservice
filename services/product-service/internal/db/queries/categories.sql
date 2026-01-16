-- name: CreateCategory :one
INSERT INTO categories (
    id,
    parent_id,
    name,
    slug,
    description,
    image_url,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCategoryByIDForUpdate :one
SELECT * FROM categories
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;

-- name: GetCategoryByName :one
SELECT * FROM categories
WHERE name = $1 AND deleted_at IS NULL;

-- name: GetCategoryBySlug :one
SELECT * FROM categories
WHERE slug = $1 AND deleted_at IS NULL;

-- name: ListAllCategories :many
SELECT * FROM categories
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: ListCategories :many
SELECT * FROM categories
WHERE
    (sqlc.narg('parent_id')::uuid IS NULL OR parent_id = sqlc.narg('parent_id'))
    AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListRootCategories :many
SELECT * FROM categories
WHERE parent_id IS NULL AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListChildCategories :many
SELECT * FROM categories
WHERE parent_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateCategory :exec
UPDATE categories SET
    name = $2,
    slug = $3,
    description = $4,
    image_url = $5,
    parent_id = $6,
    updated_at = $7
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteCategory :exec
UPDATE categories SET
    deleted_at = $2,
    updated_at = $3
WHERE id = $1;
