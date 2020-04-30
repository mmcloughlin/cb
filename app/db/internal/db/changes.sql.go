// Code generated by sqlc. DO NOT EDIT.
// source: changes.sql

package db

import (
	"context"
)

const deleteChangesCommitRange = `-- name: DeleteChangesCommitRange :exec
DELETE FROM changes
WHERE 1=1
    AND commit_index BETWEEN $1 AND $2
`

type DeleteChangesCommitRangeParams struct {
	CommitIndexMin int32
	CommitIndexMax int32
}

func (q *Queries) DeleteChangesCommitRange(ctx context.Context, arg DeleteChangesCommitRangeParams) error {
	_, err := q.exec(ctx, q.deleteChangesCommitRangeStmt, deleteChangesCommitRange, arg.CommitIndexMin, arg.CommitIndexMax)
	return err
}