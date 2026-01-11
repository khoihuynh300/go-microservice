-- name: CreateProduct :one
INSERT INTO products (
    id,
    name,
    sku,
    slug,
    description,
    category_id,
    price,
    thumbnail,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProductBySKU :one
SELECT * FROM products
WHERE sku = $1 AND deleted_at IS NULL;

-- name: GetProductBySlug :one
SELECT * FROM products
WHERE slug = $1 AND deleted_at IS NULL;

-- name: ListProducts :many
SELECT * FROM products
WHERE
    (sqlc.narg('category_id')::uuid IS NULL OR category_id = sqlc.narg('category_id'))
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountProducts :one
SELECT COUNT(*) FROM products
WHERE 
    (sqlc.narg('category_id')::uuid IS NULL OR category_id = sqlc.narg('category_id'))
    AND deleted_at IS NULL;

-- name: SearchProducts :many
SELECT * FROM products
WHERE deleted_at IS NULL
    AND (
        name ILIKE '%' || $1 || '%' 
        OR description ILIKE '%' || $1 || '%'
        OR sku ILIKE '%' || $1 || '%'
    )
ORDER BY 
    CASE 
        WHEN name ILIKE $1 || '%' THEN 1
        WHEN name ILIKE '%' || $1 || '%' THEN 2
        ELSE 3 
    END,
    created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSearchProducts :one
SELECT COUNT(*) FROM products
WHERE deleted_at IS NULL
    AND (
        name ILIKE '%' || $1 || '%' 
        OR description ILIKE '%' || $1 || '%'
        OR sku ILIKE '%' || $1 || '%'
    );

-- name: UpdateProduct :one
UPDATE products SET
    name = COALESCE(sqlc.narg('name'), name),
    sku = COALESCE(sqlc.narg('sku'), sku),
    slug = COALESCE(sqlc.narg('slug'), slug),
    description = COALESCE(sqlc.narg('description'), description),
    category_id = COALESCE(sqlc.narg('category_id'), category_id),
    price = COALESCE(sqlc.narg('price'), price),
    thumbnail = COALESCE(sqlc.narg('thumbnail'), thumbnail),
    updated_at = sqlc.narg('updated_at')
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteProduct :execrows
UPDATE products SET
    deleted_at = $2,
    updated_at = $3
WHERE id = $1;

-- name: ListProductsByIDs :many
SELECT * FROM products
WHERE id = ANY($1::uuid[]) AND deleted_at IS NULL
ORDER BY created_at DESC;
