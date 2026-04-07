-- name: CreateUser :execresult
INSERT INTO users (
	email,
	first_name,
	last_name,
	password_hash,
	middle_name
) VALUES (
	?,
	?,
	?,
	?,
	?
);

-- name: GetLoginUser :one
SELECT id, password_hash
FROM users
WHERE email = ?;

-- name: GetUser :one
SELECT email, first_name, last_name, middle_name
FROM users
WHERE id = ?;

-- name: UpdateUser :exec
UPDATE users
SET first_name = COALESCE(sqlc.narg('first_name'), first_name),
	last_name = COALESCE(sqlc.narg('last_name'), last_name),
	middle_name = COALESCE(?, middle_name)
WHERE id = ?;

-- name: UpdateLoginUser :exec
UPDATE users
SET email = COALESCE(sqlc.narg('email'), email),
	password_hash = COALESCE(sqlc.narg('password_hash'), password_hash)
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;
