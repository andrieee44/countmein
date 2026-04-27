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
INNER JOIN current_member_roles AS cmr
	ON o.organization_id = cmr.organization_id
		AND cmr.role = 'owner'
SET o.name = COALESCE(sqlc.narg(name), o.name),
	o.description = COALESCE(sqlc.narg(description), o.description),
	o.requires_join_request = COALESCE(
		sqlc.narg(requires_join_request),
		o.requires_join_request
	)
WHERE o.organization_id = sqlc.arg(organization_id)
	AND cmr.member_user_id = sqlc.arg(actor_user_id);

-- name: DeleteOrganization :execrows
DELETE o
FROM organizations AS o
INNER JOIN current_member_roles AS cmr
	ON o.organization_id = cmr.organization_id
		AND cmr.role = 'owner'
WHERE o.organization_id = sqlc.arg(organization_id)
	AND cmr.member_user_id = sqlc.arg(actor_user_id);
