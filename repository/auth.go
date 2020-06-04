package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"math/rand"

	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	result := collection.FindOne(context.TODO(), bson.M{"_id": username})
	var user *UserSchema
	err := result.Decode(&user)
	if err != nil {
		return "", errors.New("Unauthorized")
	}
	if pass := GetHash(user.Password.Salt, password); pass == user.Password.Hash {
		token := service.NewJWTService().GenerateToken(username, false)
		return token, nil
	}
	return "", errors.New("Unauthorized")
}

// NewAuthRepository function
func NewAuthRepository() AuthRepository {
	client := NewDatabaseClient()
	return &authRespository{
		client,
	}
}
