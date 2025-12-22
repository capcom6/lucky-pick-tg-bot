-- +goose Up
-- +goose StatementBegin
CREATE TABLE `group_settings` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `group_id` BIGINT UNSIGNED NOT NULL,
    `key` VARCHAR(64) NOT NULL,
    `value` TEXT NULL,
    UNIQUE KEY unique_group_setting (`group_id`, `key`),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE RESTRICT
);
-- +goose StatementEnd
-- +goose StatementBegin
INSERT INTO `group_settings` (
        `group_id`,
        `key`,
        `value`
    )
SELECT `id`,
    'discussions.delay',
    `discussions_period`
FROM `groups`;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `groups` DROP COLUMN `discussions_period`;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `groups`
ADD COLUMN `discussions_period` TINYINT UNSIGNED DEFAULT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `groups`,
    `group_settings`
SET `groups`.`discussions_period` = `group_settings`.`value`
WHERE `groups`.`id` = `group_settings`.`group_id`
    AND `group_settings`.`key` = 'discussions.delay';
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE `group_settings`;
-- +goose StatementEnd