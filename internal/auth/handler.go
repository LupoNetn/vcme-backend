package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luponetn/vcme/internal/config"
	"github.com/luponetn/vcme/internal/db"
	"github.com/luponetn/vcme/internal/util"
)

type Handler struct {
	svc    Service
	config *config.Config
}

func NewHandler(svc Service, cfg *config.Config) *Handler {
	return &Handler{
		svc:    svc,
		config: cfg,
	}
}

// handler functions
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}

	arg := db.CreateUserParams{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Bio:      pgtype.Text{String: req.Bio, Valid: req.Bio != ""},
		Location: pgtype.Text{String: req.Location, Valid: req.Location != ""},
	}

	user, err := h.svc.CreateUser(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Remove password from response for security
	user.Password = ""

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) LoginUser(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.LoginUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Remove password from response for security
	user.Password = ""

	// Generate tokens
	accessToken, err := util.GenerateToken(user.ID, user.Email, h.config.JWTAccessSecret, time.Hour*24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
		return
	}

	refreshToken, err := util.GenerateToken(user.ID, user.Email, h.config.JWTRefreshSecret, time.Hour*24*7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "user logged in successfully",
		"user":         user,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}
