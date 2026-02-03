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