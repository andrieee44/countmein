DELIMITER $$

DROP PROCEDURE IF EXISTS join_organization$$

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

	IF is_member(p_organization_id, p_actor_user_id) IS NOT FALSE THEN
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
			FROM member_roles_history
			WHERE organization_id = p_organization_id
				AND role = 'owner'
			ORDER BY created_at DESC
			LIMIT 1
		),
		NOW(6),
		'user'
	);

	COMMIT;
END$$

DROP PROCEDURE IF EXISTS leave_organization$$

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

	IF is_member(p_organization_id, p_actor_user_id) IS NOT TRUE THEN
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
		och.calendar_id,
		p_actor_user_id,
		FALSE,
		NOW(6)
	FROM organization_calendars_history AS och
	INNER JOIN calendars AS c
		ON och.calendar_id = c.calendar_id
	WHERE och.organization_id = p_organization_id
		AND c.owner_user_id = p_actor_user_id
		AND och.added = TRUE
		AND och.created_at = (
			SELECT MAX(created_at)
			FROM organization_calendars_history AS och2
			WHERE och2.calendar_id = och.calendar_id
		);

	IF NOT EXISTS (
		SELECT 1
		FROM organization_members_history AS och
		WHERE organization_id = p_organization_id
			AND added = TRUE
			AND created_at = (
				SELECT MAX(created_at)
				FROM organization_members_history AS och2
				WHERE och.organization_id = och2.organization_id
					AND och.member_user_id = och2.member_user_id
			)
		LIMIT 1
	) THEN
		DELETE FROM organizations
		WHERE organization_id = p_organization_id;

		COMMIT;
		LEAVE main;
	END IF;

	IF NOT EXISTS (
		SELECT 1
		FROM member_roles_history AS mrh
		WHERE mrh.organization_id = p_organization_id
			AND mrh.role = 'owner'
			AND mrh.created_at = (
				SELECT MAX(created_at)
				FROM member_roles_history AS mrh2
				WHERE mrh2.organization_id = mrh.organization_id
					AND mrh2.member_user_id = mrh.member_user_id
			)
	) THEN
		INSERT INTO member_roles_history (
			organization_id,
			member_user_id,
			owner_user_id,
			created_at,
			role
		) SELECT
			p_organization_id,
			omh.member_user_id,
			omh.member_user_id,
			NOW(6),
			'owner'
		FROM organization_members_history AS omh
		INNER JOIN member_roles_history AS mrh
			ON omh.organization_id = mrh.organization_id
				AND omh.member_user_id = mrh.member_user_id
		WHERE omh.organization_id = p_organization_id
			AND omh.added = TRUE
			AND omh.created_at = (
				SELECT MAX(created_at)
				FROM organization_members_history AS omh2
				WHERE omh.organization_id = omh2.organization_id
					AND omh.member_user_id = omh2.member_user_id
			)
			AND mrh.created_at = (
				SELECT MAX(created_at)
				FROM member_roles_history AS mrh2
				WHERE mrh.organization_id = mrh2.organization_id
					AND mrh.member_user_id = mrh2.member_user_id
			)
		ORDER BY
			FIELD(mrh.role, 'admin', 'user') ASC,
    		omh.created_at ASC
		LIMIT 1;

		COMMIT;
		LEAVE main;
	END IF;

	COMMIT;
END main$$

DELIMITER ;
