-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reservations (
    id TEXT PRIMARY KEY,
    court_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'confirmed',
    reserved_from TIMESTAMPTZ NOT NULL,
    reserved_to TIMESTAMPTZ NOT NULL,
    reserved_by TEXT NOT NULL,
    cancelled_by TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (reserved_from < reserved_to)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reservations;
-- +goose StatementEnd

