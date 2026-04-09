-- name: CreateCalendarCode :execresult
INSERT INTO calendars_access_codes (calendar_id, code, expires_at)
VALUES (?, ?, ?);

-- name: GetCalendarCodeMetadata :one
SELECT calendar_id, code, expires_at
FROM calendars_access_codes
WHERE id = ?;

-- name: GetCalendarCodeCalendarID :one
SELECT calendar_id
FROM calendars_access_codes
WHERE id = ?;

-- name: GetCalendarCodes :many
SELECT id
FROM calendars_access_codes
WHERE calendar_id = ?;

-- name: GetCalendarCodeFromCode :one
SELECT id, calendar_id, expires_at
FROM calendars_access_codes
WHERE code = ?;

-- name: DeleteCalendarCode :exec
DELETE FROM calendars_access_codes
WHERE id = ?;
