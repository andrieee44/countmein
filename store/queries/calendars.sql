-- name: CreateCalendar :execresult
INSERT INTO calendars (owner_id, name, ical, members_only, description)
VALUES (?, ?, ?, ?, ?);

-- name: GetCalendar :one
SELECT owner_id, name, ical, members_only, description
FROM calendars
WHERE id = ?
	AND (
		NOT members_only
		OR owner_id = sqlc.arg('user_id')
		OR EXISTS (
			SELECT 1
			FROM users_calendars
			WHERE calendar_id = calendars.id
				AND user_id = sqlc.arg('user_id')
		)
	);

-- name: GetOwnedCalendars :many
SELECT id
FROM calendars
WHERE owner_id = ?;

-- name: GetCalendarICal :one
SELECT ical
FROM calendars
WHERE id = ?;

-- name: GetCalendarOwner :one
SELECT owner_id
FROM calendars
WHERE id = ?;

-- name: ReplaceCalendar :exec
UPDATE calendars
SET ical = ?
WHERE id = ?;

-- name: UpdateMetadataCalendar :exec
UPDATE calendars
SET name = COALESCE(sqlc.narg('name'), name),
	members_only = COALESCE(sqlc.narg('members_only'), members_only),
	description = COALESCE(?, description)
WHERE id = ?;

-- name: DeleteCalendar :exec
DELETE FROM calendars
WHERE id = ?;
