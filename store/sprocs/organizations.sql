DELIMITER $$

DROP PROCEDURE IF EXISTS create_organization$$

CREATE PROCEDURE create_organization(
	IN p_actor_user_id         BIGINT,
	IN p_name                  VARCHAR(255),
	IN p_requires_join_request BOOLEAN,
	IN p_description           TEXT,
	OUT out_organization_id    BIGINT
) BEGIN
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		ROLLBACK;
		RESIGNAL;
	END;

	START TRANSACTION;

	INSERT INTO organizations (
		name,
		requires_join_request,
		created_at,
		description
	) VALUES (
		p_name,
		p_requires_join_request,
		NOW(6),
		p_description
	);

	SET out_organization_id = LAST_INSERT_ID();

	INSERT INTO organization_members_history (
		organization_id,
		member_user_id,
		added,
		created_at
	) VALUES (
		out_organization_id,
		p_actor_user_id,
		TRUE,
		NOW(6)
	);

	INSERT INTO member_roles_history (
		organization_id,
		member_user_id,
		owner_user_id,
		created_at,
		role
	) VALUES (
		out_organization_id,
		p_actor_user_id,
		p_actor_user_id,
		NOW(6),
		'owner'
	);

	COMMIT;
END$$

DELIMITER ;
