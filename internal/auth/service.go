package auth

import (
	"context"
	"time"

	"github.com/luponetn/vcme/internal/config"
	"github.com/luponetn/vcme/internal/db"
	"github.com/luponetn/vcme/internal/util"
)

type Service interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	LoginUser(ctx context.Context, email string, password string) (db.User, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
}

type Svc struct {
	queries *db.Queries
	config  *config.Config
}

func NewSvc(queries *db.Queries, config *config.Config) Service {
	return &Svc{queries: queries, config: config}
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

func (s *Svc) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := util.VerifyToken(refreshToken, s.config.JWTRefreshSecret)
	if err != nil {
		return "", "", err
	}

	//generate new refresh token and access token for user
	newAccessToken, err := util.GenerateToken(claims.UserID, claims.Email, s.config.JWTAccessSecret, time.Hour*24)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := util.GenerateToken(claims.UserID, claims.Email, s.config.JWTRefreshSecret, time.Hour*24*7)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
