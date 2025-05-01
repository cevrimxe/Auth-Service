// @title Auth API
// @version 1.0
// @description Kullanıcı giriş/çıkış işlemleri için API.
// @host localhost:8080
// @BasePath /
// @schemes http
package main

import (
	"github.com/cevrimxe/auth-service/database"

	_ "github.com/cevrimxe/auth-service/docs"
	"github.com/cevrimxe/auth-service/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	database.ConnectDB()
	server := gin.Default()
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	routes.RegisterRoutes(server)
	server.Run(":8080") //localhost:8080

}
