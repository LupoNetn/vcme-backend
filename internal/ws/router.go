package ws

import (
	"github.com/gin-gonic/gin"
	"github.com/luponetn/vcme/internal/config"
)

func RegisterWSRoutes(r *gin.Engine, manager *Manager, cfg *config.Config) {
	wsGroup := r.Group("/ws")
	{
		wsGroup.GET("/", manager.ServeWS)
	}
}
