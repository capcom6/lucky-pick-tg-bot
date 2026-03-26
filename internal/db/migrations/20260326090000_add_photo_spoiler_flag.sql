-- +goose Up
-- +goose StatementBegin
ALTER TABLE `giveaways`
ADD COLUMN `photo_has_spoiler` BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `giveaways` DROP COLUMN `photo_has_spoiler`;
-- +goose StatementEnd
