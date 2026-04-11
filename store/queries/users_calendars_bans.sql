-- name: GetCalendarBans :many
SELECT user_id
FROM users_calendars_bans
WHERE calendar_id = ?
	AND (
		expires_at IS NULL
		OR expires_at > NOW()
	);

-- name: IsMemberBanned :one
SELECT reason, expires_at, NOW(6) AS db_time
FROM users_calendars_bans
WHERE user_id = ?
	AND calendar_id = ?;

-- name: BanCalendarMember :exec
INSERT INTO users_calendars_bans (user_id, calendar_id, reason, expires_at)
VALUES (?, ?, ?, ?);

-- name: UnbanCalendarMember :exec
DELETE FROM users_calendars_bans
WHERE user_id = ?
	AND calendar_id = ?;
