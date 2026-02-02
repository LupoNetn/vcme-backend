-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS call_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
    participant_count INT NOT NULL DEFAULT 0,
    duration INT NOT NULL DEFAULT 0, -- Duration in seconds
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS call_logs;
-- +goose StatementEnd

