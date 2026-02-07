package ws

import (
	"github.com/gin-gonic/gin"
	"github.com/luponetn/vcme/internal/config"
	"github.com/luponetn/vcme/internal/middleware"
)


func RegisterWSRoutes(r *gin.Engine, manager *Manager, cfg *config.Config) {
	wsGroup := r.Group("/ws")
	wsGroup.Use(middleware.AuthMiddleware(cfg))
	{
		wsGroup.GET("/", manager.ServeWS)
	}
}