CREATE TABLE users (
	id            INT            NOT NULL AUTO_INCREMENT,
	email         VARCHAR(320)   NOT NULL,
	first_name    VARCHAR(255)   NOT NULL,
	last_name     VARCHAR(255)   NOT NULL,
	password_hash VARBINARY(255) NOT NULL,
	middle_name   VARCHAR(255),

	CONSTRAINT PRIMARY KEY (id),
	CONSTRAINT uq_users_email UNIQUE (email)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE users_sessions (
	id         BINARY(16)   NOT NULL,
	user_id    INT          NOT NULL,
	email      VARCHAR(320) NOT NULL,
	expires_at DATETIME(6)  NOT NULL,

	CONSTRAINT PRIMARY KEY (id),

	CONSTRAINT fk_users_sessions_user_id
		FOREIGN KEY (user_id)
		REFERENCES users(id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE calendars (
	id          INT          NOT NULL AUTO_INCREMENT,
	owner_id    INT          NOT NULL,
	name        VARCHAR(255) NOT NULL,
	ical_data   LONGBLOB     NOT NULL,
	description TEXT,

	CONSTRAINT PRIMARY KEY (id),

	CONSTRAINT fk_calendars_owner_id
		FOREIGN KEY (owner_id)
		REFERENCES users(id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE calendars_access_codes (
	id          INT          NOT NULL AUTO_INCREMENT,
	calendar_id INT          NOT NULL,
	code        VARCHAR(64)  NOT NULL,
	expires_at  DATETIME(6),

	CONSTRAINT PRIMARY KEY (id),
	CONSTRAINT uq_calendars_access_codes_code UNIQUE (code),

	CONSTRAINT fk_calendars_access_codes_calendar_id
		FOREIGN KEY (calendar_id)
		REFERENCES calendars(id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE users_external_calendars (
	user_id     INT NOT NULL,
	calendar_id INT NOT NULL,

	CONSTRAINT PRIMARY KEY (user_id, calendar_id),

	CONSTRAINT fk_users_external_calendars_user_id
		FOREIGN KEY (user_id)
		REFERENCES users(id)
		ON DELETE CASCADE,

	CONSTRAINT fk_users_external_calendars_calendar_id
		FOREIGN KEY (calendar_id)
		REFERENCES calendars(id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE online_calendars (
	id          INT           NOT NULL AUTO_INCREMENT,
	owner_id    INT           NOT NULL,
	name        VARCHAR(255)  NOT NULL,
	url         VARCHAR(2048) NOT NULL,
	description TEXT,

	CONSTRAINT PRIMARY KEY (id),

	CONSTRAINT fk_online_calendars_owner_id
		FOREIGN KEY (owner_id)
		REFERENCES users(id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE users_external_online_calendars (
	user_id            INT NOT NULL,
	online_calendar_id INT NOT NULL,

	CONSTRAINT PRIMARY KEY (user_id, online_calendar_id),

	CONSTRAINT fk_users_external_online_calendars_user_id
		FOREIGN KEY (user_id)
		REFERENCES users(id)
		ON DELETE CASCADE,

	CONSTRAINT fk_users_external_online_calendars_online_calendar_id
		FOREIGN KEY (online_calendar_id)
		REFERENCES online_calendars(id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
