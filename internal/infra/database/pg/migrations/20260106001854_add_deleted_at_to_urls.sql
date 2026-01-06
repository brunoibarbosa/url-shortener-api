-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_urls_deleted_at ON urls(deleted_at) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_urls_deleted_at;
ALTER TABLE urls DROP COLUMN deleted_at;
-- +goose StatementEnd
