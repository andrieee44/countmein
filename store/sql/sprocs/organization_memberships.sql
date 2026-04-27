DELIMITER $$

CREATE PROCEDURE join_organization(
	IN p_actor_user_id   BIGINT,
	IN p_organization_id BIGINT
) BEGIN
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		ROLLBACK;
		RESIGNAL;
	END;

	START TRANSACTION;

	IF (
		SELECT requires_join_request
		FROM organizations
		WHERE organization_id = p_organization_id
	) IS NOT FALSE THEN
		SIGNAL SQLSTATE '45000'
			SET MESSAGE_TEXT = 'resource not found or access denied';
	END IF;

	IF EXISTS (
		SELECT 1
		FROM current_memberships
		WHERE organization_id = p_organization_id
			AND member_user_id = p_actor_user_id
	) THEN
		SIGNAL SQLSTATE '45000'
			SET MESSAGE_TEXT = 'membership already exists';
	END IF;

	INSERT INTO organization_members_history (
		organization_id,
		member_user_id,
		added,
		created_at
	) VALUES (
		p_organization_id,
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
		p_organization_id,
		p_actor_user_id,
		(
			SELECT member_user_id
			FROM current_member_roles
			WHERE organization_id = p_organization_id
				AND role = 'owner'
			LIMIT 1
		),
		NOW(6),
		'user'
	);

	COMMIT;
END$$

CREATE PROCEDURE leave_organization(
	IN p_actor_user_id   BIGINT,
	IN p_organization_id BIGINT
) main: BEGIN
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		ROLLBACK;
		RESIGNAL;
	END;

	START TRANSACTION;

	IF NOT EXISTS (
		SELECT 1
		FROM current_memberships
		WHERE organization_id = p_organization_id
			AND member_user_id = p_actor_user_id
	) THEN
		SIGNAL SQLSTATE '45000'
			SET MESSAGE_TEXT = 'membership doesn''t exist';
	END IF;

	INSERT INTO organization_members_history (
		organization_id,
		member_user_id,
		added,
		created_at
	) VALUES (
		p_organization_id,
		p_actor_user_id,
		FALSE,
		NOW(6)
	);

	INSERT INTO organization_calendars_history (
		organization_id,
		calendar_id,
		admin_user_id,
		added,
		created_at
	) SELECT
		p_organization_id,
		coc.calendar_id,
		p_actor_user_id,
		FALSE,
		NOW(6)
	FROM current_organization_calendars AS coc
	INNER JOIN calendars AS c
		ON coc.calendar_id = c.calendar_id
	WHERE coc.organization_id = p_organization_id
		AND c.owner_user_id = p_actor_user_id;

	IF NOT EXISTS (
		SELECT 1
		FROM current_memberships
		WHERE organization_id = p_organization_id
	) THEN
		DELETE FROM organizations
		WHERE organization_id = p_organization_id;

		COMMIT;
		LEAVE main;
	END IF;

	IF NOT EXISTS (
		SELECT 1
		FROM current_member_roles
		WHERE organization_id = p_organization_id
			AND role = 'owner'
	) THEN
		INSERT INTO member_roles_history (
			organization_id,
			member_user_id,
			owner_user_id,
			created_at,
			role
		) SELECT
			p_organization_id,
			cmr.member_user_id,
			cmr.member_user_id,
			NOW(6),
			'owner'
		FROM current_member_roles AS cmr
		WHERE cmr.organization_id = p_organization_id
		ORDER BY
			FIELD(cmr.role, 'admin', 'user') ASC,
			(
				SELECT created_at
				FROM organization_members_history
				WHERE organization_id = p_organization_id
					AND member_user_id = cmr.member_user_id
				ORDER BY created_at ASC
				LIMIT 1
			) ASC
		LIMIT 1;

		COMMIT;
		LEAVE main;
	END IF;

	COMMIT;
END main$$

DELIMITER ;
