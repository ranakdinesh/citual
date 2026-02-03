-- name: CreateRole :one
INSERT INTO roles (
    id,
    tenant_id,
    name,
    code,
    description
) VALUES (
             $1, $2, $3, $4, $5
         )
RETURNING *;

-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1 AND tenant_id = $2;

-- name: GetRoleByCode :one
SELECT * FROM roles
WHERE code = $1 AND tenant_id = $2;

-- name: ListRoles :many
SELECT * FROM roles
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2;

-- name: ListUserRoles :many
SELECT r.* FROM roles r
                    JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1 AND tenant_id = $2;
