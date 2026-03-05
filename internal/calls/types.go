package call

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCallRequestParams struct {
	Title       string      `json:"title"`
	Description pgtype.Text `json:"description"`
	CallLink    string      `json:"call_link"`
	HostID      uuid.UUID   `json:"host_id"`
}

type EndCallParams struct {
	UserID           uuid.UUID   `json:"user_id"`
	CallID           uuid.UUID   `json:"call_id"`
	Participant      pgtype.Text `json:"participant"`
	Type             pgtype.Text `json:"type"`
	Time             pgtype.Text `json:"time"`
	CallTitle        pgtype.Text `json:"call_title"`
	Duration         string      `json:"duration"`
	ParticipantCount int32       `json:"participant_count"`
}
