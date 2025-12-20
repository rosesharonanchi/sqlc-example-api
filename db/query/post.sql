-- name: CreatePost :one
INSERT INTO posts (
  title,
  content,
  user_id
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetPost :one
SELECT * FROM posts
WHERE id = $1 LIMIT 1;

-- name: ListPosts :many
SELECT * FROM posts
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdatePost :one
UPDATE posts
SET title = $2, content = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1;