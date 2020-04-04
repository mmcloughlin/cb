// Code generated by sqlc. DO NOT EDIT.
// source: modules.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const insertModule = `-- name: InsertModule :exec
INSERT INTO modules (
    uuid,
    path,
    version
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT DO NOTHING
`

type InsertModuleParams struct {
	UUID    uuid.UUID
	Path    string
	Version string
}

func (q *Queries) InsertModule(ctx context.Context, arg InsertModuleParams) error {
	_, err := q.exec(ctx, q.insertModuleStmt, insertModule, arg.UUID, arg.Path, arg.Version)
	return err
}

const module = `-- name: Module :one
SELECT uuid, path, version FROM modules
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) Module(ctx context.Context, uuid uuid.UUID) (Module, error) {
	row := q.queryRow(ctx, q.moduleStmt, module, uuid)
	var i Module
	err := row.Scan(&i.UUID, &i.Path, &i.Version)
	return i, err
}

const modules = `-- name: Modules :many
SELECT uuid, path, version FROM modules
`

func (q *Queries) Modules(ctx context.Context) ([]Module, error) {
	rows, err := q.query(ctx, q.modulesStmt, modules)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Module
	for rows.Next() {
		var i Module
		if err := rows.Scan(&i.UUID, &i.Path, &i.Version); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
