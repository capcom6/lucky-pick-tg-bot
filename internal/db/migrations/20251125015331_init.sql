-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    telegram_user_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(255) NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NULL,
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    INDEX idx_username (username)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE groups (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    telegram_group_id BIGINT NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE giveaways (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    group_id BIGINT UNSIGNED NOT NULL,
    admin_user_id BIGINT UNSIGNED NOT NULL,
    photo_file_id VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    publish_date DATETIME NOT NULL,
    application_end_date DATETIME NOT NULL,
    results_date DATETIME NOT NULL,
    is_anonymous BOOLEAN DEFAULT FALSE,
    status ENUM('scheduled', 'active', 'finished', 'cancelled') DEFAULT 'scheduled',
    telegram_message_id BIGINT NULL,
    is_pinned BOOLEAN DEFAULT FALSE,
    winner_user_id BIGINT UNSIGNED NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE RESTRICT,
    FOREIGN KEY (admin_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (winner_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    INDEX idx_status (status),
    INDEX idx_publish_date (publish_date),
    INDEX idx_application_end_date (application_end_date),
    INDEX idx_results_date (results_date),
    INDEX idx_group_status (group_id, status)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE giveaway_participants (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    giveaway_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (giveaway_id) REFERENCES giveaways(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    UNIQUE KEY unique_participation (giveaway_id, user_id),
    INDEX idx_user_id (user_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE action_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    giveaway_id BIGINT UNSIGNED NULL,
    user_id BIGINT UNSIGNED NULL,
    action_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (giveaway_id) REFERENCES giveaways(id) ON DELETE
    SET NULL,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE
    SET NULL,
        INDEX idx_action_type (action_type),
        INDEX idx_created_at (created_at)
);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP TABLE action_logs;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE giveaway_participants;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE giveaways;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE groups;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd