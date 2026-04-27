-- name: GetCalendarWrites :many
SELECT calendar_write_event_id
FROM calendar_writes_history AS cwh
INNER JOIN calendars AS c
	ON cwh.calendar_id = c.calendar_id
WHERE c.calendar_id = ?
	AND c.owner_user_id = sqlc.arg(actor_user_id);

-- name: WriteCalendar :execlastid
INSERT INTO calendar_writes_history (
	calendar_id,
	writer_user_id,
	ical_encrypted,
	created_at
) SELECT
	sqlc.arg(calendar_id),
	sqlc.arg(actor_user_id2),
	AES_ENCRYPT(sqlc.arg(ical), sqlc.arg(AES_SECRET_KEY)),
	NOW(6)
WHERE EXISTS (
	SELECT 1
	FROM calendars AS c
	WHERE c.calendar_id = sqlc.arg(calendar_id)
		AND c.owner_user_id = sqlc.arg(actor_user_id)
) OR EXISTS (
	SELECT 1
	FROM current_organization_calendars AS coc
	INNER JOIN current_member_roles AS cmr
		ON coc.organization_id = cmr.organization_id
			AND cmr.member_user_id = sqlc.arg(actor_user_id)
			AND cmr.role IN ('owner', 'admin')
	WHERE coc.calendar_id = sqlc.arg(calendar_id)
);

-- name: GetCalendarWrite :one
SELECT cwh.calendar_id,
	writer_user_id,
	AES_DECRYPT(ical_encrypted, sqlc.arg(AES_SECRET_KEY)) AS ical,
	created_at
FROM calendar_writes_history AS cwh
INNER JOIN calendars AS c
	ON cwh.calendar_id = c.calendar_id
WHERE cwh.calendar_write_event_id = ?
	AND c.owner_user_id = sqlc.arg(actor_user_id);
