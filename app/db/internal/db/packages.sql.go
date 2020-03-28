// Code generated by sqlc. DO NOT EDIT.
// source: packages.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const insertPkg = `-- name: InsertPkg :exec
INSERT INTO packages (
    uuid,
    module_uuid,
    relative_path
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT DO NOTHING
`

type InsertPkgParams struct {
	UUID         uuid.UUID
	ModuleUUID   uuid.UUID
	RelativePath string
}

func (q *Queries) InsertPkg(ctx context.Context, arg InsertPkgParams) error {
	_, err := q.db.ExecContext(ctx, insertPkg, arg.UUID, arg.ModuleUUID, arg.RelativePath)
	return err
}

const modulePkgs = `-- name: ModulePkgs :many
SELECT uuid, module_uuid, relative_path FROM packages
WHERE module_uuid = $1
`

func (q *Queries) ModulePkgs(ctx context.Context, moduleUuid uuid.UUID) ([]Package, error) {
	rows, err := q.db.QueryContext(ctx, modulePkgs, moduleUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Package
	for rows.Next() {
		var i Package
		if err := rows.Scan(&i.UUID, &i.ModuleUUID, &i.RelativePath); err != nil {
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

const pkg = `-- name: Pkg :one
SELECT uuid, module_uuid, relative_path FROM packages
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) Pkg(ctx context.Context, uuid uuid.UUID) (Package, error) {
	row := q.db.QueryRowContext(ctx, pkg, uuid)
	var i Package
	err := row.Scan(&i.UUID, &i.ModuleUUID, &i.RelativePath)
	return i, err
}
