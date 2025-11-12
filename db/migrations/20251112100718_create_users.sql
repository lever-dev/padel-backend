-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    phone_number  TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT now(),
    last_login_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_phone ON users(phone_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_phone;

DROP TABLE IF EXISTS users;
-- +goose StatementEnd
