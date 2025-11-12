-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    nickname TEXT NOT NULL,
    password TEXT NOT NULL,
    phone_number  TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT now(),
    last_login_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_phone ON users(phone_number);
CREATE UNIQUE INDEX idx_users_nickname ON users(nickname);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_nickname;

DROP TABLE IF EXISTS users;
-- +goose StatementEnd
