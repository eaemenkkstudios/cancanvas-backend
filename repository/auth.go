package repository

import (
	"context"
	"errors"
	"math/rand"

	"golang.org/x/crypto/bcrypt"

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
	hash, err := bcrypt.GenerateFromPassword([]byte(salt+password), 20)
	if err != nil {
		return ""
	}
	return string(hash[:])
}

// ValidatePassword function
func ValidatePassword(password string, salt string, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+password)) == nil
}

func (db *authRespository) Login(username string, password string) (string, error) {
	collection := db.client.Database(Database).Collection(CollectionUsers)
	user := collection.FindOne(context.TODO(), bson.M{"_id": username})
	var u *UserSchema
	err := user.Decode(&u)
	if err != nil {
		return "", err
	}
	if valid := ValidatePassword(password, u.Password.Salt, u.Password.Hash); valid {
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
