-- name: ToggleShareUserCalendar :exec
CALL toggle_share_user_calendar(
	sqlc.arg(actor_user_id),
	sqlc.arg(organization_id),
	sqlc.arg(calendar_id)
);

-- name: GetOrganizationCalendars :many
SELECT calendar_id
FROM current_organization_calendars AS coc
INNER JOIN current_memberships AS cm
	ON coc.organization_id = cm.organization_id
WHERE coc.organization_id = ?
	AND cm.member_user_id = sqlc.arg(actor_user_id);
