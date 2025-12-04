
-- name: CreatePost :one
INSERT INTO "post" ("user_id", "title", "content")
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM "post"
WHERE "id" = $1 LIMIT 1;

-- name: ListAllPosts :many
SELECT * FROM "post"
ORDER BY "created_at" DESC;

-- name: DeletePost :exec
DELETE FROM "post"
WHERE "id" = $1 AND "user_id" = $2;

-- name: UpdatePostContent :one
UPDATE "post"
SET "content" = $3
WHERE "id" = $1 AND "user_id" = $2
RETURNING *;