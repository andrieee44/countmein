-- name: GetSubscribedCalendars :many
SELECT calendar_id
FROM users_calendars
WHERE user_id = ?;

-- name: SubscribeToCalendar :exec
INSERT INTO users_calendars (user_id, calendar_id, color)
VALUES (?, ?, ?);

-- name: UnsubscribeFromCalendar :exec
DELETE FROM users_calendars
WHERE user_id = ?
	AND calendar_id = ?;

-- name: GetCalendarMembers :many
SELECT user_id
FROM users_calendars
WHERE calendar_id = ?;

-- name: GetSubscribedMetadata :one
SELECT color
FROM users_calendars
WHERE user_id = ?
	AND calendar_id = ?;

-- name: UpdateSubscribedMetadata :exec
UPDATE users_calendars
SET color = COALESCE(sqlc.narg('color'), color)
WHERE user_id = ?
	AND calendar_id = ?;
