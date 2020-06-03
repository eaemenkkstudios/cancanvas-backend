package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AuthRepository interface
type AuthRepository interface {
	Login(username string, password string) (string, error)
}

type authRespository struct {
	client *mongo.Client
}

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-$")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GetSalt function
func GetSalt() string {
	return randSeq(20)
}

// GetHash function
func GetHash(salt string, password string) string {
	hash := sha256.Sum256([]byte(salt + password))
	return string(hash[:])
}

func (db *authRespository) Login(username string, password string) (string, error) {
	collection := db.client.Database(Database).Collection(CollectionUsers)
	user := collection.FindOne(context.TODO(), bson.M{"_id": username})
	var u *UserSchema
	err := user.Decode(&u)
	if err != nil {
		return "", err
	}
	if pass := GetHash(u.Password.Salt, password); pass == u.Password.Hash {
		token := service.NewJWTService().GenerateToken(username, false)
		return token, nil
	}
	return "", errors.New("Unauthorized")
}

// NewAuthRepository function
func NewAuthRepository() AuthRepository {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found.")
	}
	MONGODB := os.Getenv("MONGODB_URL")

	clientOptions := options.Client().ApplyURI(MONGODB)
	clientOptions = clientOptions.SetMaxPoolSize(50)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database Connection Succeeded!")

	return &authRespository{
		client,
	}
}
