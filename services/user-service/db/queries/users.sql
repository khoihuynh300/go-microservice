-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    id, email, hashed_password, full_name, phone, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET
    full_name = $2,
    phone = $3,
    avatar_url = $4,
    date_of_birth = $5,
    gender = $6,
    updated_at = $7
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $2, updated_at = $3
WHERE id = $1;

-- name: UpdateUserStatus :exec
UPDATE users
SET status = $2, updated_at = $3
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE status = $1;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = $2
WHERE id = $1;
