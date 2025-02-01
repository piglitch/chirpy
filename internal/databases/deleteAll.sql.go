// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: deleteAll.sql

package databases

import (
	"context"
)

const deleteUser = `-- name: DeleteUser :one
 DELETE FROM users
RETURNING id, created_at, updated_at, email, hashed_password
`

func (q *Queries) DeleteUser(ctx context.Context) (User, error) {
	row := q.db.QueryRowContext(ctx, deleteUser)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
	)
	return i, err
}
