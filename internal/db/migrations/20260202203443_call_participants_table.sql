-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS call_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id UUID NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'participant',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    -- Ensure a user can't be in the same call twice with the same role simultaneously 
    -- (though they might join/leave, so we don't necessarily want a unique constraint on just call_id/user_id if we want history)
    -- For now, a simple unique constraint on call_id and user_id might be desired if only one record per user per call is allowed.
    UNIQUE(call_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS call_participants;
-- +goose StatementEnd

