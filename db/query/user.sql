-- name: CreateUser :one
INSERT INTO users(username, email , hashed_password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUseryByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET username = $2 , hashed_password = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

