
-- name: CreateUser :one
INSERT INTO "user" ("user_name", "password_hash")
VALUES ($1, $2)
RETURNING "id", "user_name", "created_at";

-- name: GetUserByUsername :one
SELECT * FROM "user"
WHERE "user_name" = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT "id", "user_name", "created_at" FROM "user"
WHERE "id" = $1 LIMIT 1;

-- name: ListUsers :many
SELECT "id", "user_name", "created_at" FROM "user"
ORDER BY "created_at" DESC;