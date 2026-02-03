-- name: CreateUser :one
INSERT INTO users (
    id,
    tenant_id,
    first_name,
    last_name,
    email,
    mobile,
    password_hash,
    is_super_admin,
    is_active
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9
         )
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 AND tenant_id = $2;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: UpdateUser :exec
UPDATE users
SET
    first_name = $2,
    last_name = $3,
    mobile = $4,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $5;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1 AND tenant_id = $2;
