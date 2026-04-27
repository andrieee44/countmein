-- name: JoinOrganization :exec
CALL join_organization(
	sqlc.arg(actor_user_id),
	sqlc.arg(organization_id)
);

-- name: LeaveOrganization :exec
CALL leave_organization(
	sqlc.arg(actor_user_id),
	sqlc.arg(organization_id)
);
