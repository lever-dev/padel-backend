-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS courts (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_organization_id ON courts (organization_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_organization_id;

DROP TABLE IF EXISTS courts;
-- +goose StatementEnd