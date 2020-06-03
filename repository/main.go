package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database and Collection names
const (
	Database        = "cancanvas"
	CollectionUsers = "users"
)

var client *mongo.Client

// NewDatabaseClient function
func NewDatabaseClient() *mongo.Client {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found.")
	}
	MONGODB := os.Getenv("MONGODB_URL")

	clientOptions := options.Client().ApplyURI(MONGODB)
	clientOptions = clientOptions.SetMaxPoolSize(50)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if client == nil {
		client, err = mongo.Connect(ctx, clientOptions)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database Connection Succeeded!")

	return client
}
