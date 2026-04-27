DELIMITER $$

DROP FUNCTION IF EXISTS is_member$$

CREATE FUNCTION is_member(
	p_organization_id BIGINT,
	p_user_id         BIGINT
) RETURNS BOOLEAN
READS SQL DATA
BEGIN
	RETURN EXISTS (
		SELECT 1
		FROM organization_members_history AS omh
		WHERE organization_id = p_organization_id
			AND member_user_id = p_user_id
			AND added = TRUE
			AND created_at = (
				SELECT MAX(created_at)
				FROM organization_members_history AS omh2
				WHERE omh.organization_id = omh2.organization_id
					AND omh.member_user_id = omh2.member_user_id
			)
	);
END$$

DELIMITER ;
