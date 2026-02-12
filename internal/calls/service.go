package call

import (
	"context"

	"github.com/google/uuid"
	"github.com/luponetn/vcme/internal/db"
)

type Service interface {
	CreateCallLink(ctx context.Context, arg db.CreateCallLinkParams) (db.Call, error)
	ListAllCallsByID(ctx context.Context, hostID uuid.UUID) ([]db.Call, error)
	ListAllCalls(ctx context.Context) ([]db.Call, error)
	GetCallByLink(ctx context.Context, link string) (db.GetCallByLinkRow, error)
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

func (s *Svc) ListAllCallsByID(ctx context.Context, hostID uuid.UUID) ([]db.Call, error) {
	return s.queries.ListAllCallsByID(ctx, hostID)
}

func (s *Svc) ListAllCalls(ctx context.Context) ([]db.Call, error) {
	return s.queries.ListAllCalls(ctx)
}

func (s *Svc) GetCallByLink(ctx context.Context, link string) (db.GetCallByLinkRow, error) {
	return s.queries.GetCallByLink(ctx, link)
}
