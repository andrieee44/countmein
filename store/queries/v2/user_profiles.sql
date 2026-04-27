-- name: GetUserOwnedCalendars :many
SELECT calendar_id
FROM calendars
WHERE owner_user_id = sqlc.arg(actor_user_id);

-- name: GetUserOrganizations :many
SELECT organization_id
FROM current_memberships
WHERE member_user_id = sqlc.arg(actor_user_id);

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
