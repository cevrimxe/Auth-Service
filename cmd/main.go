package main

import (
	"github.com/cevrimxe/auth-service/database"
	"github.com/cevrimxe/auth-service/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()
	server := gin.Default()
	routes.RegisterRoutes(server)
	server.Run(":8080") //localhost:8080

}
