package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository Interface
type UserRepository interface {
	Save(user *model.NewUser) *model.User
	FindAll() []*model.User
}

// Database and Collection names
const (
	DATABASE   = "cancanvas"
	COLLECTION = "users"
)

type database struct {
	client *mongo.Client
}

func (db *database) Save(user *model.NewUser) *model.User {
	collection := db.client.Database(DATABASE).Collection(COLLECTION)
	result, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}
	oid, _ := result.InsertedID.(primitive.ObjectID)
	return &model.User{
		ID:       oid.Hex(),
		Email:    user.Email,
		Name:     user.Name,
		Password: user.Password,
	}
}

func (db *database) FindAll() []*model.User {
	collection := db.client.Database(DATABASE).Collection(COLLECTION)
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())
	var users []*model.User
	for cursor.Next(context.TODO()) {
		var u *model.User
		err := cursor.Decode(&u)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}
	return users
}

// New Database Client
func New() UserRepository {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found.")
	}
	MONGODB := os.Getenv("MONGODB_URL")

	clientOptions := options.Client().ApplyURI(MONGODB)
	clientOptions = clientOptions.SetMaxPoolSize(50)

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database Connection Succeeded!")

	return &database{
		client,
	}
}
