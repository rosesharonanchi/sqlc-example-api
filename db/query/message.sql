-- -- name: CreateMessage :one
-- INSERT INTO message (thread, user_id, content)
-- VALUES ($1, $2, $3)
-- RETURNING *;

-- -- name: GetMessageByID :one
-- SELECT * FROM message
-- WHERE id = $1;

-- -- name: GetMessagesByThread :many
-- SELECT * FROM message
-- WHERE thread = $1
-- ORDER BY created_at DESC;


-- -- name: DeleteMessage :exec
-- DELETE FROM message
-- WHERE id = $1 AND user_id = $2;

-- -- name: UpdateMessage :one
-- UPDATE message 
-- SET content = $3
-- WHERE id = $1 AND user_id = $2
-- RETURNING *;

