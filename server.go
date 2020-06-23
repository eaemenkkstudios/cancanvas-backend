package main

import (
	"os"

	"github.com/eaemenkkstudios/cancanvas-backend/middleware"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	server := gin.Default()

	server.GET("/", middleware.PlaygroundHandler())
	server.GET("/query", middleware.GraphQLHandler())
	server.POST("/query", middleware.GraphQLHandler())
	server.GET("/cancel", middleware.CancelHandler())
	server.GET("/return", middleware.ResultHandler())
	server.GET("/resetpassword", middleware.ResetPasswordHandler())

	server.Run(":" + port)
}
