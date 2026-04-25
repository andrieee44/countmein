DELIMITER $$

DROP PROCEDURE IF EXISTS create_calendar$$

CREATE PROCEDURE create_calendar(
	IN p_actor_user_id  BIGINT,
	IN p_name           VARCHAR(255),
	IN p_ical           LONGBLOB,
	IN p_description    TEXT,
	IN p_AES_SECRET_KEY BINARY(32),
	OUT out_calendar_id BIGINT
) BEGIN
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		ROLLBACK;
		RESIGNAL;
	END;

	START TRANSACTION;

	INSERT INTO calendars (
		owner_user_id,
		name,
		description
	) VALUES (
		p_actor_user_id,
		p_name,
		p_description
	);

	SET out_calendar_id = LAST_INSERT_ID();

	INSERT INTO calendar_writes_history (
		calendar_id,
		writer_user_id,
		ical_encrypted,
		created_at
	) VALUES (
		out_calendar_id,
		p_actor_user_id,
		AES_ENCRYPT(p_ical, p_AES_SECRET_KEY),
		NOW(6)
	);

	COMMIT;
END$$

DELIMITER ;
