-- name: CreateUserLabel :execlastid
INSERT INTO users_labels (user_id, name, color)
VALUES (sqlc.arg(actor_user_id), ?, ?);

-- name: GetUserLabel :one
SELECT name, color
FROM users_labels
WHERE user_label_id = ?
	AND user_id = sqlc.arg(actor_user_id);

-- name: UpdateUserLabel :execrows
UPDATE users_labels
SET name = COALESCE(sqlc.narg(name), name),
	color = COALESCE(sqlc.narg(color), color)
WHERE user_label_id = ?
	AND user_id = sqlc.arg(actor_user_id);

-- name: DeleteUserLabel :execrows
DELETE FROM users_labels
WHERE user_label_id = ?
	AND user_id = sqlc.arg(actor_user_id);

-- name: AttachUserLabel :exec
INSERT INTO calendars_users_labels (calendar_id, user_label_id)
SELECT ?, sqlc.arg(user_label_id)
WHERE EXISTS (
	SELECT 1
	FROM users_labels AS ul
	WHERE ul.user_label_id = sqlc.arg(user_label_id)
	AND ul.user_id = sqlc.arg(actor_user_id)
);

-- name: DetachUserLabel :execrows
DELETE cul
FROM calendars_users_labels AS cul
INNER JOIN users_labels AS ul
	ON cul.user_label_id = ul.user_label_id
WHERE cul.calendar_id = ?
	AND ul.user_label_id = ?
	AND ul.user_id = sqlc.arg(actor_user_id);
