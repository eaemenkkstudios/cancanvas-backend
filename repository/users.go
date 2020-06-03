package repository

import (
	"context"
	"log"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepository Interface
type UserRepository interface {
	Save(user *model.NewUser) *model.User
	FindAll() []*model.User
}

type userRepository struct {
	client *mongo.Client
}

type password struct {
	Hash string `json:"hash"`
	Salt string `json:"salt"`
}

// UserSchema struct
type UserSchema struct {
	Nickname  string   `json:"nickname" bson:"_id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Artist    bool     `json:"artist"`
	Gallery   []string `json:"gallery"`
	Followers int      `json:"followers"`
	Following []string `json:"following"`
	Password  password `json:"password"`
}

func (db *userRepository) Save(user *model.NewUser) *model.User {
	collection := db.client.Database(Database).Collection(CollectionUsers)
	salt := GetSalt()
	_, err := collection.InsertOne(context.TODO(), &UserSchema{
		Email:     user.Email,
		Nickname:  user.Nickname,
		Artist:    user.Artist,
		Name:      user.Name,
		Gallery:   make([]string, 0),
		Following: make([]string, 0),
		Followers: 0,
		Password: password{
			Hash: GetHash(salt, user.Password),
			Salt: salt,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	return &model.User{
		Email:    user.Email,
		Artist:   user.Artist,
		Name:     user.Name,
		Nickname: user.Nickname,
	}
}

func (db *userRepository) FindAll() []*model.User {
	collection := db.client.Database(Database).Collection(CollectionUsers)
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

// NewUserRepository function
func NewUserRepository() UserRepository {
	client := NewDatabaseClient()
	return &userRepository{
		client,
	}
}
