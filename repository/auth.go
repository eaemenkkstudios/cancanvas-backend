package repository

import (
	"context"
	"crypto/sha256"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuthRepository interface
type AuthRepository interface {
	Login(username, password string) (*model.Login, error)
	ResetPassword(sender, hash, newPassword string) (bool, error)
	SendForgotPasswordEmail(user string) (bool, error)
}

type authRespository struct {
	client *mongo.Database
}

// Password struct
type Password struct {
	Hash string `json:"hash"`
	Salt string `json:"salt"`
}

var letters = []byte("*_#$-&abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) string {
	rand.Seed(time.Now().Unix())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GetSalt function
func GetSalt() string {
	return randSeq(0x20)
}

// GetHash function
func GetHash(salt, password string) string {
	multiplier := 2
	hash := sha256.Sum256([]byte(salt + password))
	hashString := make([]byte, len(hash)*multiplier)
	for i, b := range hash {
		hashString[i*multiplier] = byte(letters[int(b>>4)%len(letters)])
		hashString[i*multiplier+1] = byte(letters[int(b&0xf)%len(letters)])
	}
	return string(hashString[:])
}

// GeneratePassword function
func GeneratePassword(password string) Password {
	salt := GetSalt()
	return Password{
		Salt: salt,
		Hash: GetHash(salt, password),
	}
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

func (db *authRespository) ResetPassword(sender, hash, newPassword string) (bool, error) {
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return false, errors.New("Could not reset password")
	}
	if hash != u.Password.Hash {
		println(hash)
		println(u.Password.Hash)
		return false, errors.New("Invalid token")
	}
	password := GeneratePassword(newPassword)
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"password": password},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *authRespository) SendForgotPasswordEmail(user string) (bool, error) {
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": user})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return false, errors.New("User not found")
	}
	token := service.NewJWTService().GenerateResetPasswordToken(user, u.Password.Hash)
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
