-- +goose Up
-- +goose StatementBegin
CREATE TABLE `giveaway_discussions` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `giveaway_id` BIGINT UNSIGNED NOT NULL,
    `user_id` BIGINT UNSIGNED,
    `text` TEXT NOT NULL,
    `telegram_message_id` BIGINT,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_giveaway_created (giveaway_id, created_at DESC),
    FOREIGN KEY (giveaway_id) REFERENCES giveaways(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `groups`
ADD COLUMN `discussions_period` TINYINT UNSIGNED DEFAULT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `groups`
MODIFY COLUMN `is_active` BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `users`
MODIFY COLUMN `is_active` BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `users`
MODIFY COLUMN `is_active` BOOLEAN DEFAULT TRUE;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `groups`
MODIFY COLUMN `is_active` BOOLEAN DEFAULT TRUE;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `groups` DROP COLUMN `discussions_period`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE `giveaway_discussions`;
-- +goose StatementEnd