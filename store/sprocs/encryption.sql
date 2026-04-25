DELIMITER $$

DROP PROCEDURE IF EXISTS SSELECT$$

CREATE PROCEDURE SSELECT(
	IN p_table    VARCHAR(255),
	IN p_database VARCHAR(255),
	IN p_where    VARCHAR(1000)
) BEGIN
	DECLARE sql_query TEXT;

	SELECT CONCAT(
		'SELECT ',
		GROUP_CONCAT('TO_BASE64(`', COLUMN_NAME, '`) AS `', COLUMN_NAME, '`' SEPARATOR ', '),
		' FROM `', p_database, '`.`', p_table, '`',
		IF (p_where IS NOT NULL, CONCAT(' WHERE ', p_where), '')
	)
	INTO sql_query
	FROM INFORMATION_SCHEMA.COLUMNS
	WHERE TABLE_SCHEMA = p_database
		AND TABLE_NAME = p_table
	GROUP BY TABLE_NAME;

	SET @query = sql_query;
	PREPARE stmt FROM @query;
	EXECUTE stmt;
	DEALLOCATE PREPARE stmt;
END$$

DELIMITER ;
