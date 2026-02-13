-- +goose Up
-- +goose StatementBegin
ALTER TABLE call_logs
ALTER COLUMN time TYPE TEXT USING time::TEXT,
ALTER COLUMN duration TYPE TEXT USING duration::TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE call_logs
ALTER COLUMN time TYPE INTEGER USING time::INTEGER,
ALTER COLUMN duration TYPE INTEGER USING duration::INTEGER;
-- +goose StatementEnd
