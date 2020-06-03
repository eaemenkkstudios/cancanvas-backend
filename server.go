package main

import (
	"fmt"
	"os"

	"github.com/eaemenkkstudios/cancanvas-backend/http"
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

	server.GET("/", http.PlaygroundHandler())
	server.POST("/query", http.GraphQLHandler())

	server.Run(":" + port)
}
