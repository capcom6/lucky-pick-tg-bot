-- +goose Up
-- +goose StatementBegin
ALTER TABLE `users`
MODIFY COLUMN `registered_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `groups`
MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    MODIFY COLUMN `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `giveaways`
MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    MODIFY COLUMN `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `giveaway_participants`
MODIFY COLUMN `joined_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `action_logs`
MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `action_logs`
MODIFY COLUMN `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `giveaway_participants`
MODIFY COLUMN `joined_at` DATETIME DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `giveaways`
MODIFY COLUMN `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    MODIFY COLUMN `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `groups`
MODIFY COLUMN `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    MODIFY COLUMN `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `users`
MODIFY COLUMN `registered_at` DATETIME DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd