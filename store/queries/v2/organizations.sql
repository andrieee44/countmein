-- name: CreateOrganization :exec
CALL create_organization(
	sqlc.arg(actor_user_id),
	sqlc.arg(name),
	sqlc.arg(requires_join_request),
	sqlc.arg(description),
	@out_organization_id
);

-- name: GetOrganization :one
SELECT name, requires_join_request, created_at, description
FROM organizations
WHERE organization_id = ?;

-- name: GetOrganizations :many
SELECT organization_id
FROM organizations;

-- name: UpdateOrganization :execrows
UPDATE organizations AS o
INNER JOIN member_roles_history AS mrh
	ON o.organization_id = mrh.organization_id
SET o.name = COALESCE(sqlc.narg(name), o.name),
	o.description = COALESCE(sqlc.narg(description), o.description),
	o.requires_join_request = COALESCE(
		sqlc.narg(requires_join_request),
		o.requires_join_request
	)
WHERE o.organization_id = ?
	AND mrh.member_user_id = sqlc.arg(actor_user_id)
	AND mrh.role = 'owner'
	AND mrh.created_at = (
		SELECT MAX(created_at)
		FROM member_roles_history AS mrh2
		WHERE o.organization_id = mrh2.organization_id
			AND mrh.member_user_id = mrh2.member_user_id
	);

-- name: DeleteOrganization :execrows
DELETE o
FROM organizations AS o
INNER JOIN member_roles_history AS mrh
	ON o.organization_id = mrh.organization_id
WHERE o.organization_id = ?
	AND mrh.member_user_id = sqlc.arg(actor_user_id)
	AND mrh.role = 'owner'
	AND mrh.created_at = (
		SELECT MAX(created_at)
		FROM member_roles_history AS mrh2
		WHERE o.organization_id = mrh2.organization_id
			AND mrh.member_user_id = mrh2.member_user_id
	);
