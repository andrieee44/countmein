-- name: CreateUserSession :exec
INSERT INTO users_sessions (id, user_id, expires_at)
VALUES (
	?,
	sqlc.arg(user_id),
	FROM_UNIXTIME(UNIX_TIMESTAMP() + sqlc.arg(ttl_seconds))
);

-- name: GetUserSession :one
SELECT user_id, expires_at, NOW(6) AS db_time
FROM users_sessions
WHERE id = sqlc.arg(session_id);

-- name: RevokeUserSession :exec
DELETE FROM users_sessions
WHERE id = sqlc.arg(session_id);

-- name: RevokeAllUserSessions :exec
DELETE FROM users_sessions
WHERE user_id = sqlc.arg(actor_user_id);
