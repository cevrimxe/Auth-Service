package routes

import (
	"github.com/cevrimxe/auth-service/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	server.POST("/signup", signup)
	server.POST("/login", login)
	server.GET("/verify", verifyEmail)

	authenticated := server.Group("/")
	authenticated.Use(middlewares.Authenticate)
	authenticated.GET("/me", getMe) // get infos of loggedin user // need authorization

	// server.POST("/forgot-password",forgotPassword)
	// server.POST("/reset-password",resetPassword) // reset password with mailed reset link
	// server.POST("/logout",logout) // logout // need authorization
	// server.PUT("/me",update) // update loggedin user // need authorization
	// server.GET("/admin/users",) // get users // admin role needed
}
