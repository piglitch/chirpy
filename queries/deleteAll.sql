-- name: DeleteUser :one
 DELETE FROM users
RETURNING *;