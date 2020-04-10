-- name: Task :one
SELECT * FROM tasks
WHERE uuid = $1 LIMIT 1;

-- name: WorkerTasksWithStatus :many
SELECT
    *
FROM
    tasks
WHERE 1=1
    AND worker=sqlc.arg(worker)
    AND status = ANY (sqlc.arg(statuses)::task_status[])
;

-- name: CreateTask :one
INSERT INTO tasks (
    uuid,
    worker,
    commit_sha,
    type,
    target_uuid,
    status,
    last_status_update
)VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    'created',
    NOW()
)
RETURNING *
;

-- name: TransitionTaskStatus :one
UPDATE
    tasks
SET
    status = CASE WHEN status = sqlc.arg(status_from) THEN sqlc.arg(status_to) ELSE status END,
    last_status_update = CASE WHEN status = sqlc.arg(status_from) THEN NOW() ELSE last_status_update END
WHERE 1=1
    AND uuid=sqlc.arg(uuid)
RETURNING
    status
;
