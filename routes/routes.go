package routes

import (
	"github.com/cevrimxe/auth-service/handlers"
	"github.com/cevrimxe/auth-service/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine, userHandler *handlers.UserHandler) {
	server.POST("/signup", userHandler.Signup)
	server.POST("/login", userHandler.Login)
	server.GET("/verify", userHandler.VerifyEmail)

	authenticated := server.Group("/")
	authenticated.Use(middlewares.Authenticate)
	authenticated.GET("/me", userHandler.GetMe)
	authenticated.PUT("/me", userHandler.UpdateMe)
	authenticated.PUT("/change-password", userHandler.ChangePassword)
	authenticated.GET("/admin/users", userHandler.GetUsers)

	server.POST("/forgot-password", userHandler.ForgetPassword)
	server.POST("/reset-password", userHandler.ResetPassword)
}
