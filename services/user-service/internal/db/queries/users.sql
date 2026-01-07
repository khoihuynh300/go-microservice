-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    id, email, hashed_password, full_name, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING
    id,
    email,
    full_name,
    status,
    created_at,
    updated_at;

-- name: UpdateUser :execrows
UPDATE users
SET
    full_name = $2,
    phone = $3,
    date_of_birth = $4,
    gender = $5,
    updated_at = $6
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserAvatar :execrows
UPDATE users
SET avatar_url = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: VerifyUserEmail :execrows
UPDATE users
SET email_verified_at = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserPassword :execrows
UPDATE users
SET hashed_password = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserStatus :execrows
UPDATE users
SET status = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :execrows
UPDATE users
SET deleted_at = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;
