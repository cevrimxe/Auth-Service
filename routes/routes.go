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
	authenticated.GET("/me", getMe)                       // get infos of loggedin user // need authorization
	authenticated.PUT("/me", updateMe)                    // update loggedin user // need authorization
	authenticated.PUT("/change-password", changePassword) // update loggedin user password. I break it with put /me for security // need authorization
	authenticated.GET("/admin/users", getUsers)           // get users // admin role needed

	server.POST("/forgot-password", forgetPassword)
	server.POST("/reset-password", resetPassword) // reset password with mailed reset link

	// server.PUT("/me",update)
	// server.GET("/admin/users",)

	// authenticated.POST("/logout", logout) // server.POST("/logout",logout) // logout // need authorization  // frontend can delete token basicly so dont need endpoint for this
}
