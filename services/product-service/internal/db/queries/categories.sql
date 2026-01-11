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

-- name: UpdateCategory :one
UPDATE categories SET
    name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    description = COALESCE(sqlc.narg('description'), description),
    image_url = COALESCE(sqlc.narg('image_url'), image_url),
    parent_id = COALESCE(sqlc.narg('parent_id'), parent_id),
    updated_at = sqlc.narg('updated_at')
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCategory :execrows
UPDATE categories SET
    deleted_at = $2,
    updated_at = $3
WHERE id = $1;
