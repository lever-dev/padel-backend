-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_organizations_name_city ON organizations (LOWER(city), LOWER(name));

CREATE INDEX idx_organizations_city ON organizations (city);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_organizations_name_city;

DROP INDEX IF EXISTS idx_organizations_city;

DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd