-- name: GetPassword :one
SELECT hashed_password FROM users
WHERE email = $1;