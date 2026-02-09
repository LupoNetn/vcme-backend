package call

import (
	"context"

	"github.com/google/uuid"
	"github.com/luponetn/vcme/internal/db"
)

type Service interface {
	CreateCallLink(ctx context.Context, arg db.CreateCallLinkParams) (db.Call, error)
	ListCallsByHostID(ctx context.Context, hostID uuid.UUID) ([]db.Call, error)
}

type Svc struct {
	queries *db.Queries
}

func NewSvc(q *db.Queries) Service {
	return &Svc{
		queries: q,
	}
}

//call functions which imlements service

func (s *Svc) CreateCallLink(ctx context.Context, arg db.CreateCallLinkParams) (db.Call, error) {
	return s.queries.CreateCallLink(ctx, arg)
}

func (s *Svc) ListCallsByHostID(ctx context.Context, hostID uuid.UUID) ([]db.Call, error) {
	return s.queries.ListCallsByHostID(ctx, hostID)
}
