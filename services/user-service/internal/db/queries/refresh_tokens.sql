-- name: GetRefreshTokenByTokenHash :one
SELECT id, user_id, token_hash, expires_at, created_at
FROM refresh_tokens
WHERE token_hash = $1;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (
    id, user_id, token_hash, expires_at, created_at
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: DeleteRefreshTokenByID :execrows
DELETE FROM refresh_tokens
WHERE id = $1;
