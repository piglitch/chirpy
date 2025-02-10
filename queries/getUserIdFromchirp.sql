-- name: UserIdFromChirp :one
SELECT user_id FROM chirps
WHERE id = $1;