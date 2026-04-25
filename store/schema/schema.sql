CREATE TABLE users (
	user_id       BIGINT         NOT NULL AUTO_INCREMENT,
	email         VARCHAR(320)   NOT NULL,
	first_name    VARCHAR(255)   NOT NULL,
	last_name     VARCHAR(255)   NOT NULL,
	password_hash VARBINARY(255) NOT NULL,
	middle_name   VARCHAR(255),

	CONSTRAINT PRIMARY KEY (user_id),
	CONSTRAINT uq_users_email UNIQUE (email)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE users_sessions (
	session_hash BINARY(32)   NOT NULL,
	user_id      BIGINT       NOT NULL,
	expires_at   DATETIME(6)  NOT NULL,

	CONSTRAINT PRIMARY KEY (session_hash),

	CONSTRAINT fk_users_sessions_user_id
		FOREIGN KEY (user_id)
		REFERENCES users(user_id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE users_labels (
	user_label_id BIGINT       NOT NULL AUTO_INCREMENT,
	user_id       BIGINT       NOT NULL,
	name          VARCHAR(255) NOT NULL,
	color         CHAR(6)      NOT NULL,

	CONSTRAINT PRIMARY KEY (user_label_id),

	CONSTRAINT fk_users_labels_user_id
		FOREIGN KEY (user_id)
		REFERENCES users(user_id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE calendars (
	calendar_id    BIGINT       NOT NULL AUTO_INCREMENT,
	owner_user_id  BIGINT       NOT NULL,
	name           VARCHAR(255) NOT NULL,
	description    TEXT,

	CONSTRAINT PRIMARY KEY (calendar_id),

	CONSTRAINT fk_calendars_owner_user_id
		FOREIGN KEY (owner_user_id)
		REFERENCES users(user_id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE calendar_writes_history (
	calendar_write_event_id BIGINT      NOT NULL AUTO_INCREMENT,
	calendar_id             BIGINT      NOT NULL,
	writer_user_id          BIGINT,
	ical_encrypted          LONGBLOB    NOT NULL,
	created_at              DATETIME(6) NOT NULL,

	CONSTRAINT PRIMARY KEY (calendar_write_event_id),

	CONSTRAINT fk_calendar_writes_history_calendar_id
		FOREIGN KEY (calendar_id)
		REFERENCES calendars(calendar_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_calendar_writes_history_writer_user_id
		FOREIGN KEY (writer_user_id)
		REFERENCES users(user_id)
		ON DELETE SET NULL
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE calendars_users_labels (
	calendar_id   BIGINT NOT NULL,
	user_label_id BIGINT NOT NULL,

	CONSTRAINT PRIMARY KEY (calendar_id, user_label_id),

	CONSTRAINT fk_calendars_users_labels_calendar_id
		FOREIGN KEY (calendar_id)
		REFERENCES calendars(calendar_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_calendars_users_labels_user_label_id
		FOREIGN KEY (user_label_id)
		REFERENCES users_labels(user_label_id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE organizations (
	organization_id       BIGINT       NOT NULL AUTO_INCREMENT,
	name                  VARCHAR(255) NOT NULL,
	requires_join_request BOOLEAN      NOT NULL,
	created_at            DATETIME(6)  NOT NULL,
	description           TEXT,

	CONSTRAINT PRIMARY KEY (organization_id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE organization_members_history (
	organization_member_event_id BIGINT      NOT NULL AUTO_INCREMENT,
	organization_id              BIGINT      NOT NULL,
	member_user_id               BIGINT      NOT NULL,
	added                        BOOLEAN     NOT NULL,
	created_at                   DATETIME(6) NOT NULL,

	CONSTRAINT PRIMARY KEY (organization_member_event_id),

	CONSTRAINT fk_organization_members_history_organization_id
		FOREIGN KEY (organization_id)
		REFERENCES organizations(organization_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_organization_members_history_member_user_id
		FOREIGN KEY (member_user_id)
		REFERENCES users(user_id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE organization_calendars_history (
	organization_calendar_event_id BIGINT      NOT NULL AUTO_INCREMENT,
	organization_id                BIGINT      NOT NULL,
	calendar_id                    BIGINT      NOT NULL,
	admin_user_id                  BIGINT,
	added                          BOOLEAN     NOT NULL,
	created_at                     DATETIME(6) NOT NULL,

	CONSTRAINT PRIMARY KEY (organization_calendar_event_id),

	CONSTRAINT fk_organization_calendars_history_organization_id
		FOREIGN KEY (organization_id)
		REFERENCES organizations(organization_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_organization_calendars_history_calendar_id
		FOREIGN KEY (calendar_id)
		REFERENCES calendars(calendar_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_organization_calendars_history_admin_user_id
		FOREIGN KEY (admin_user_id)
		REFERENCES users(user_id)
		ON DELETE SET NULL
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE join_prompts_history (
	join_prompt_event_id BIGINT      NOT NULL AUTO_INCREMENT,
	organization_id      BIGINT      NOT NULL,
	owner_user_id        BIGINT,
	prompt               TEXT        NOT NULL,
	created_at           DATETIME(6) NOT NULL,

	CONSTRAINT PRIMARY KEY (join_prompt_event_id),

	CONSTRAINT fk_join_prompts_history_organization_id
		FOREIGN KEY (organization_id)
		REFERENCES organizations(organization_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_join_prompts_history_owner_user_id
		FOREIGN KEY (owner_user_id)
		REFERENCES users(user_id)
		ON DELETE SET NULL
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE join_responses_history (
	join_response_event_id BIGINT      NOT NULL AUTO_INCREMENT,
	join_prompt_event_id   BIGINT      NOT NULL,
	user_id                BIGINT      NOT NULL,
	response               TEXT        NOT NULL,
	created_at             DATETIME(6) NOT NULL,

	CONSTRAINT PRIMARY KEY (join_response_event_id),

	CONSTRAINT fk_join_responses_history_join_prompt_event_id
		FOREIGN KEY (join_prompt_event_id)
		REFERENCES join_prompts_history(join_prompt_event_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_join_responses_history_user_id
		FOREIGN KEY (user_id)
		REFERENCES users(user_id)
		ON DELETE CASCADE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE join_requests_history (
	join_request_event_id  BIGINT      NOT NULL AUTO_INCREMENT,
	join_response_event_id BIGINT      NOT NULL,
	admin_user_id          BIGINT,
	created_at             DATETIME(6) NOT NULL,

	status ENUM(
		'open',
		'retracted',
		'denied',
		'accepted'
	) NOT NULL,

	CONSTRAINT PRIMARY KEY (join_request_event_id),

	CONSTRAINT fk_join_requests_history_join_response_event_id
		FOREIGN KEY (join_response_event_id)
		REFERENCES join_responses_history(join_response_event_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_join_requests_history_admin_user_id
		FOREIGN KEY (admin_user_id)
		REFERENCES users(user_id)
		ON DELETE SET NULL
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE member_roles_history (
	member_role_event_id BIGINT      NOT NULL AUTO_INCREMENT,
	organization_id      BIGINT      NOT NULL,
	member_user_id       BIGINT      NOT NULL,
	owner_user_id        BIGINT,
	created_at           DATETIME(6) NOT NULL,

	role ENUM(
		'owner',
		'admin',
		'user'
	) NOT NULL,

	CONSTRAINT PRIMARY KEY (member_role_event_id),

	CONSTRAINT fk_member_roles_history_organization_id
		FOREIGN KEY (organization_id)
		REFERENCES organizations(organization_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_member_roles_history_member_user_id
		FOREIGN KEY (member_user_id)
		REFERENCES users(user_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_member_roles_history_owner_user_id
		FOREIGN KEY (owner_user_id)
		REFERENCES users(user_id)
		ON DELETE SET NULL
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE member_moderations_history (
	member_moderation_event_id BIGINT       NOT NULL AUTO_INCREMENT,
	organization_id            BIGINT       NOT NULL,
	member_user_id             BIGINT       NOT NULL,
	admin_user_id              BIGINT,
	created_at                 DATETIME(6)  NOT NULL,
	reason                     VARCHAR(255) NOT NULL,
	expires_at                 DATETIME(6),

	action ENUM(
		'ban',
		'unban'
	) NOT NULL,

	CONSTRAINT PRIMARY KEY (member_moderation_event_id),

	CONSTRAINT fk_member_moderations_history_organization_id
		FOREIGN KEY (organization_id)
		REFERENCES organizations(organization_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_member_moderations_history_member_user_id
		FOREIGN KEY (member_user_id)
		REFERENCES users(user_id)
		ON DELETE CASCADE,

	CONSTRAINT fk_member_moderations_history_admin_user_id
		FOREIGN KEY (admin_user_id)
		REFERENCES users(user_id)
		ON DELETE SET NULL
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
