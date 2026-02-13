-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS call_logs
ADD COLUMN user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS call_logs
-- +goose StatementEnd

