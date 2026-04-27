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
SELECT owner_user_id,
	name,
	description,
	AES_DECRYPT(cwh.ical_encrypted, sqlc.arg(AES_SECRET_KEY)) AS ical,
	cwh.created_at AS updated_at
FROM calendars AS c
INNER JOIN calendar_writes_history AS cwh
	ON c.calendar_id = cwh.calendar_id
WHERE c.calendar_id = sqlc.arg(calendar_id)
	AND cwh.created_at = (
		SELECT MAX(created_at)
		FROM calendar_writes_history AS cwh2
		WHERE cwh2.calendar_id = cwh.calendar_id
	) AND (
		c.owner_user_id = sqlc.arg(actor_user_id)
		OR EXISTS (
			SELECT 1
			FROM current_organization_calendars AS coc
			INNER JOIN current_memberships AS cm
				ON coc.organization_id = cm.organization_id
					AND cm.member_user_id = sqlc.arg(actor_user_id)
			WHERE coc.calendar_id = c.calendar_id
		)
	);

-- name: UpdateCalendar :execrows
UPDATE calendars
SET name = COALESCE(sqlc.narg(name), name),
	description = COALESCE(sqlc.narg(description), description)
WHERE calendar_id = ?
	AND owner_user_id = sqlc.arg(actor_user_id);

-- name: DeleteCalendar :execrows
DELETE FROM calendars
WHERE calendar_id = ?
	AND owner_user_id = sqlc.arg(actor_user_id);
