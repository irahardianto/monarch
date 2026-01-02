-- name: CreateProject :one
INSERT INTO projects (path) VALUES ($1) RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE path = $1 LIMIT 1;

-- name: CreateTask :one
INSERT INTO tasks (project_id, title, status) VALUES ($1, $2, $3) RETURNING *;

-- name: ListTasks :many
SELECT * FROM tasks WHERE project_id = $1;

-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1 LIMIT 1;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = $2 WHERE id = $1;

-- name: IncrementTaskAttempt :one
UPDATE tasks SET attempt_count = attempt_count + 1 WHERE id = $1 RETURNING attempt_count;
