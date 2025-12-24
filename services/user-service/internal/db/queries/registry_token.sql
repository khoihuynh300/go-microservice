-- name: CreateRegistryToken :exec
INSERT INTO registry_tokens (
    id, user_id, token_hash, expires_at, created_at
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: GetActiveRegistryToken :one
SELECT id, user_id, token_hash, expires_at, created_at, used_at, invalidated_at FROM registry_tokens
WHERE token_hash = $1 AND used_at IS NULL AND invalidated_at IS NULL AND expires_at > NOW();

-- name: MarkRegistryTokenAsUsed :exec
UPDATE registry_tokens
SET used_at = NOW()
WHERE token_hash = $1 AND used_at IS NULL AND invalidated_at IS NULL AND expires_at > NOW();

-- name: InvalidateRegistryTokens :exec
UPDATE registry_tokens
SET invalidated_at = NOW()
WHERE user_id = $1 AND used_at IS NULL AND invalidated_at IS NULL AND expires_at > NOW();
