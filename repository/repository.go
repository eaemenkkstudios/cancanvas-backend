package repository

import (
	"context"
	"log"
	"os"
	"time"

	//
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database and Collection names
const (
	Database        = "cancanvas"
	CollectionUsers = "users"
	CollectionChats = "chats"
	CollectionPosts = "posts"
)

// NewDatabaseClient function
func newDatabaseClient() *mongo.Client {
	MONGODB := os.Getenv("MONGODB_URL")

	clientOptions := options.Client().ApplyURI(MONGODB)
	clientOptions = clientOptions.SetMaxPoolSize(50)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	return client
}
