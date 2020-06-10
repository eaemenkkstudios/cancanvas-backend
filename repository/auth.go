package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"math/rand"
	"os"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthRepository interface
type AuthRepository interface {
	Login(username, password string) (*model.Login, error)
	SendForgotPasswordEmail(user string) (bool, error)
}

type authRespository struct {
	client *mongo.Database
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
func GetHash(salt, password string) string {
	hash := sha256.Sum256([]byte(salt + password))
	return string(hash[:])
}

func (db *authRespository) Login(username, password string) (*model.Login, error) {
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": username})
	var user *UserSchema
	err := result.Decode(&user)
	if err != nil {
		return nil, errors.New("Unauthorized")
	}
	if pass := GetHash(user.Password.Salt, password); pass == user.Password.Hash {
		token := service.NewJWTService().GenerateToken(username, false)
		if user.First {
			collection.UpdateOne(context.TODO(), bson.M{"_id": username}, bson.M{
				"$set": bson.M{"first": false},
			})
		}
		return &model.Login{
			Token: token,
			First: user.First,
		}, nil
	}
	return nil, errors.New("Unauthorized")
}

func (db *authRespository) SendForgotPasswordEmail(user string) (bool, error) {
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": user})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return false, errors.New("User not found")
	}
	token := service.NewJWTService().GenerateResetPasswordToken(user)
	serverURL := os.Getenv("SERVER_URL")
	err = service.NewMailerService().SendMail(u.Email, "Hi,\n\n"+
		"A password reset was request to the account associated with the cdias900@gmail.com email address, click the link bellow to change your password:\n"+
		serverURL+"/resetpassword?token="+token+"\n"+
		"If you didn't request this change, please ignore this email.\n\n"+
		"Cancavas Team\n",
	)
	if err != nil {
		return false, errors.New("Could not send email")
	}
	return true, nil
}

// NewAuthRepository function
func NewAuthRepository() AuthRepository {
	client := newDatabaseClient()
	return &authRespository{
		client,
	}
}
