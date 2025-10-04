-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS user_profiles (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT,
    avatar_url TEXT,
    phone TEXT,
    birth_date DATE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_profiles;
-- +goose StatementEnd
