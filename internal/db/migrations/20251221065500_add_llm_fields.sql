-- +goose Up
-- +goose StatementBegin
ALTER TABLE `giveaways`
ADD COLUMN `original_description` TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `giveaways`
SET `original_description` = `description`;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `giveaways` DROP COLUMN `original_description`;
-- +goose StatementEnd