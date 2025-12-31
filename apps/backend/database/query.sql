-- name: CreateProject :one
INSERT INTO projects (path) VALUES ($1) RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE path = $1 LIMIT 1;

-- name: CreateTask :one
INSERT INTO tasks (project_id, title, status) VALUES ($1, $2, $3) RETURNING *;

-- name: ListTasks :many
SELECT * FROM tasks WHERE project_id = $1;
