-- name: GetUserOwnedCalendars :many
SELECT calendar_id
FROM calendars
WHERE owner_user_id = sqlc.arg(actor_user_id);

-- name: GetUserOrganizations :many
SELECT organization_id
FROM organization_members_history AS omh
WHERE omh.member_user_id = sqlc.arg(actor_user_id)
	AND omh.added = TRUE
	AND omh.created_at = (
		SELECT MAX(created_at)
		FROM organization_members_history AS omh2
		WHERE omh2.organization_id = omh.organization_id
			AND omh2.member_user_id = omh.member_user_id
	);

-- name: GetUserLabels :many
SELECT user_label_id
FROM users_labels
WHERE user_id = sqlc.arg(actor_user_id);

-- name: GetUserCalendarLabels :many
SELECT ul.user_label_id
FROM users_labels AS ul
INNER JOIN calendars_users_labels AS cul
    ON cul.user_label_id = ul.user_label_id
WHERE cul.calendar_id = ?
    AND ul.user_id = sqlc.arg(actor_user_id);

-- name: GetUserEmail :one
SELECT email
FROM users
WHERE user_id = sqlc.arg(actor_user_id)
