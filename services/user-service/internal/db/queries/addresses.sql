-- name: GetAddressByIDAndUserID :one
SELECT * FROM user_addresses
WHERE id = $1 AND user_id = $2;

-- name: ListAddressesByUserID :many
SELECT * FROM user_addresses
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateAddress :one
INSERT INTO user_addresses (
    id, user_id, address_type, full_name, phone,
    address_line1, address_line2, ward, city, country,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12
)
RETURNING *;

-- name: UpdateAddress :execrows
UPDATE user_addresses
SET
    address_type = $2,
    full_name = $3,
    phone = $4,
    address_line1 = $5,
    address_line2 = $6,
    ward = $7,
    city = $8,
    country = $9,
    updated_at = $10
WHERE id = $1;

-- name: SetDefaultAddress :execrows
UPDATE user_addresses
SET is_default = CASE WHEN id = $2 THEN TRUE ELSE FALSE END
WHERE user_id = $1;

-- name: DeleteAddress :execrows
DELETE FROM user_addresses WHERE id = $1;
