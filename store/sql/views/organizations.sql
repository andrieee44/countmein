CREATE VIEW current_memberships AS
SELECT organization_id,
	member_user_id
FROM organization_members_history AS omh
WHERE added = TRUE
	AND created_at = (
		SELECT MAX(created_at)
		FROM organization_members_history AS omh2
		WHERE omh.organization_id = omh2.organization_id
			AND omh.member_user_id = omh2.member_user_id
	);

CREATE VIEW current_member_roles AS
SELECT mrh.organization_id,
	mrh.member_user_id,
	role
FROM member_roles_history AS mrh
INNER JOIN current_memberships AS cm
	ON mrh.organization_id = cm.organization_id
		AND mrh.member_user_id = cm.member_user_id
WHERE mrh.created_at = (
		SELECT MAX(created_at)
		FROM member_roles_history AS mrh2
		WHERE mrh.organization_id = mrh2.organization_id
			AND mrh.member_user_id = mrh2.member_user_id
	);

CREATE VIEW current_organization_calendars AS
SELECT organization_id,
	calendar_id
FROM organization_calendars_history AS och
WHERE added = TRUE
	AND created_at = (
		SELECT MAX(created_at)
		FROM organization_calendars_history AS och2
		WHERE och.organization_id = och2.organization_id
			AND och.calendar_id = och2.calendar_id
	);
