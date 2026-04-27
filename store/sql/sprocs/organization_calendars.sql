DELIMITER $$

CREATE PROCEDURE toggle_share_user_calendar(
	IN p_actor_user_id   BIGINT,
	IN p_organization_id BIGINT,
	IN p_calendar_id     BIGINT
) BEGIN
	DECLARE v_is_shared BOOLEAN;
	DECLARE v_is_owner  BOOLEAN;
	DECLARE v_is_admin  BOOLEAN;

	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		ROLLBACK;
		RESIGNAL;
	END;

	START TRANSACTION;

	SELECT EXISTS (
		SELECT 1
		FROM current_organization_calendars
		WHERE organization_id = p_organization_id
		  AND calendar_id = p_calendar_id
	) INTO v_is_shared;

	SELECT EXISTS (
		SELECT 1
		FROM calendars
		WHERE calendar_id = p_calendar_id
		  AND owner_user_id = p_actor_user_id
	) INTO v_is_owner;

	SELECT EXISTS (
		SELECT 1
		FROM current_member_roles
		WHERE organization_id = p_organization_id
		  AND member_user_id = p_actor_user_id
		  AND role IN ('owner', 'admin')
	) INTO v_is_admin;

	IF NOT (
		(v_is_shared AND (v_is_owner OR v_is_admin))
		OR (NOT v_is_shared AND (v_is_owner AND v_is_admin))
	) THEN
		SIGNAL SQLSTATE '45000'
			SET MESSAGE_TEXT = 'resource not found or access denied';
	END IF;

	INSERT INTO organization_calendars_history (
		organization_id,
		calendar_id,
		admin_user_id,
		added,
		created_at
	) VALUES (
		p_organization_id,
		p_calendar_id,
		p_actor_user_id,
		NOT v_is_shared,
		NOW(6)
	);

	COMMIT;
END$$

DELIMITER ;
