-- +goose Up
-- +goose StatementBegin
CREATE TABLE `group_admins` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `group_id` BIGINT UNSIGNED NOT NULL,
    `user_id` BIGINT UNSIGNED NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE RESTRICT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    UNIQUE KEY unique_group_admin (group_id, user_id)
);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP TABLE `group_admins`;
-- +goose StatementEnd