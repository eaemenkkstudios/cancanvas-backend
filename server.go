package main

import (
	"fmt"
	"os"

	"github.com/eaemenkkstudios/cancanvas-backend/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found.")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	server := gin.Default()

	server.GET("/", middleware.PlaygroundHandler())
	server.POST("/query", middleware.GraphQLHandler())

	server.Run(":" + port)
}
