package auth

import (
	"context"

	"github.com/luponetn/vcme/internal/db"
	"github.com/luponetn/vcme/internal/util"
)

type Service interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	LoginUser(ctx context.Context, email string, password string) (db.User, error)
}

type Svc struct {
	queries *db.Queries
}

func NewSvc(queries *db.Queries) Service {
	return &Svc{queries: queries}
}

// functions implementationss
func (s *Svc) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return s.queries.CreateUser(ctx, arg)
}

func (s *Svc) LoginUser(ctx context.Context, email string, password string) (db.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return db.User{}, err
	}

	err = util.CheckPassword(password, user.Password)
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}
