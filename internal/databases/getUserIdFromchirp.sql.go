// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: getUserIdFromchirp.sql

package databases

import (
	"context"

	"github.com/google/uuid"
)

const userIdFromChirp = `-- name: UserIdFromChirp :one
SELECT user_id FROM chirps
WHERE id = $1
`

func (q *Queries) UserIdFromChirp(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, userIdFromChirp, id)
	var user_id uuid.UUID
	err := row.Scan(&user_id)
	return user_id, err
}
