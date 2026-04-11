-- name: CreateUserSession :exec
INSERT INTO users_sessions (id, user_id, email, expires_at)
VALUES (?, ?, ?, ?);

-- name: GetUserSession :one
SELECT id, user_id, email, expires_at, NOW(6) AS db_time
FROM users_sessions
WHERE id = ?;

-- name: UpdateUserSession :exec
UPDATE users_sessions
SET email = (
	SELECT email
	FROM users
	WHERE id = users_sessions.user_id
)
WHERE user_id = ?;

-- name: RevokeUserSession :exec
DELETE FROM users_sessions
WHERE id = ?;

-- name: RevokeAllUserSession :exec
DELETE FROM users_sessions
WHERE user_id = ?;
