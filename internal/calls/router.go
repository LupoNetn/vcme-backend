package call

import (
	"github.com/gin-gonic/gin"
	"github.com/luponetn/vcme/internal/config"
	"github.com/luponetn/vcme/internal/middleware"
)

func RegisterCallRoutes(r *gin.Engine, h *Handler, cfg *config.Config) {
	callGroup := r.Group("/calls")
	callGroup.Use(middleware.AuthMiddleware(cfg))

	callGroup.POST("/", h.CreateCallLink)
}
