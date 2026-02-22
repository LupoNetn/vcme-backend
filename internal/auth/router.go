package auth

import "github.com/gin-gonic/gin"

func RegisterAuthRoutes(router *gin.Engine, h *Handler) {
	authGroup := router.Group("/auth")

	authGroup.POST("/signup", h.CreateUser)
	authGroup.POST("/login", h.LoginUser)

	//create a separate router for refreshing the users token, this route does
	//not belong to the authgroup because the user woud be requesting for a new
	//token using the old token
	router.POST("/refresh", h.Refresh)
}
