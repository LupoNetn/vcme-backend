package auth

import "github.com/gin-gonic/gin"

func RegisterAuthRoutes(router *gin.Engine, h *Handler) {
	authGroup := router.Group("/auth")

	authGroup.POST("/signup", h.CreateUser)
	authGroup.POST("/login", h.LoginUser)
}
