-- name: CreateProductImage :one
INSERT INTO product_images (
    ID, product_id, image_url, position, created_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetProductImages :many
SELECT * FROM product_images
WHERE product_id = $1
ORDER BY position ASC, created_at ASC;

-- name: GetProductImagesForUpdate :many
SELECT * FROM product_images
WHERE product_id = $1
ORDER BY position ASC, created_at ASC
FOR UPDATE;

-- name: DeleteProductImage :exec
DELETE FROM product_images
WHERE id = $1;

-- name: DeleteAllProductImages :exec
DELETE FROM product_images
WHERE product_id = $1;

-- name: UpdateImagePosition :exec
UPDATE product_images SET position = $2
WHERE id = $1;
