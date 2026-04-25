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
	FROM organization_calendars_history AS och
	INNER JOIN organization_members_history AS omh
		ON och.organization_id = omh.organization_id
	INNER JOIN member_roles_history AS mrh
		ON och.organization_id = mrh.organization_id
			AND omh.member_user_id = mrh.member_user_id
	WHERE och.calendar_id = sqlc.arg(calendar_id)
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
		AND mrh.role IN ('owner', 'admin')
		AND mrh.created_at = (
			SELECT MAX(created_at)
			FROM member_roles_history AS mrh2
			WHERE mrh2.organization_id = mrh.organization_id
				AND mrh2.member_user_id = mrh.member_user_id
		)
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
