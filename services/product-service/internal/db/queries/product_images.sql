-- name: CreateProductImage :one
INSERT INTO product_images (
    product_id, image_url, position, created_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetProductImages :many
SELECT * FROM product_images
WHERE product_id = $1
ORDER BY position ASC, created_at ASC;

-- name: DeleteProductImage :execrows
DELETE FROM product_images
WHERE id = $1;

-- name: DeleteAllProductImages :execrows
DELETE FROM product_images
WHERE product_id = $1;

-- name: UpdateImagePosition :execrows
UPDATE product_images SET position = $2
WHERE id = $1;
