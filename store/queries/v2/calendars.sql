-- name: CreateCalendar :exec
CALL create_calendar(
	sqlc.arg(actor_user_id),
	sqlc.arg(name),
	sqlc.arg(ical),
	sqlc.arg(description),
	sqlc.arg(AES_SECRET_KEY),
	@out_calendar_id
);

-- name: GetCalendar :one
SELECT
	owner_user_id,
	name,
	description,
	AES_DECRYPT(cwh.ical_encrypted, sqlc.arg(AES_SECRET_KEY)) AS ical,
	cwh.created_at AS updated_at
FROM calendars AS c
INNER JOIN calendar_writes_history AS cwh
	ON c.calendar_id = cwh.calendar_id
WHERE c.calendar_id = ?
	AND cwh.created_at = (
		SELECT MAX(created_at)
		FROM calendar_writes_history AS cwh2
		WHERE cwh2.calendar_id = cwh.calendar_id
	)
	AND (
		c.owner_user_id = sqlc.arg(actor_user_id)
		OR EXISTS (
			SELECT 1
			FROM organization_calendars_history AS och
			INNER JOIN organization_members_history AS omh
				ON och.organization_id = omh.organization_id
			WHERE och.calendar_id = c.calendar_id
				AND och.added = TRUE
				AND och.created_at = (
					SELECT MAX(created_at)
					FROM organization_calendars_history AS och2
					WHERE och2.calendar_id = och.calendar_id
				)
				AND omh.member_user_id = sqlc.arg(actor_user_id)
				AND omh.added = TRUE
				AND omh.created_at = (
					SELECT MAX(created_at)
					FROM organization_members_history AS omh2
					WHERE omh2.organization_id = omh.organization_id
						AND omh2.member_user_id = omh.member_user_id
				)
		)
	);

-- name: UpdateCalendarMetadata :execrows
UPDATE calendars
SET name = COALESCE(sqlc.narg(name), name),
	description = COALESCE(sqlc.narg(description), description)
WHERE calendar_id = ?
	AND owner_user_id = sqlc.arg(actor_user_id);

-- name: DeleteCalendar :execrows
DELETE FROM calendars
WHERE calendar_id = ?
	AND owner_user_id = sqlc.arg(actor_user_id);
