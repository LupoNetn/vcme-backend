-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS call_logs 
ADD COLUMN type TEXT,
ADD COLUMN time INT,
ADD COLUMN call_title TEXT,
ADD COLUMN participant TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP COLUMN type,
DROP COLUMN time,
DROP COLUMN call_title,
DROP COLUMN participant,
-- +goose StatementEnd
