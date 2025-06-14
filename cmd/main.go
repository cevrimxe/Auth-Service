// @title Auth API
// @version 1.0
// @description Kullanıcı giriş/çıkış işlemleri için API.
// @host localhost:8080
// @BasePath /
// @schemes http
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := &http.Server{
		Addr:    ":8080",
		Handler: server,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server could not start: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forcefully shut down:", err)
	}

	if err := database.CloseDB(); err != nil {
		log.Fatal("Database connection could not be closed:", err)
	}

	log.Println("Server gracefully shut down")
}
