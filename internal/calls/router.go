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
	callGroup.GET("/", h.ListAllCalls)
	callGroup.GET("/:host_id", h.ListCallsByHostID)

	// Public search route
	r.GET("/calls/link/*link", h.GetCallByLink)
}
