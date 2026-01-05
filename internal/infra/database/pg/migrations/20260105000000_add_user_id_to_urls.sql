-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_urls_user_id ON urls(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_urls_user_id;
ALTER TABLE urls DROP COLUMN user_id;
-- +goose StatementEnd
