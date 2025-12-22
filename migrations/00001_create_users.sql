-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    phone TEXT UNIQUE NOT NULL,
    cod integer,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,

    version integer NOT NULL DEFAULT 1,
    deleted bool NOT NULL DEFAULT false,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    created_by BIGINT, 
    updated_at timestamp(0) with time zone,
    updated_by BIGINT
);

CREATE INDEX IF NOT EXISTS idx_users_cod ON users(cod);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted ON users(deleted) WHERE NOT deleted;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS citext;
-- +goose StatementEnd
