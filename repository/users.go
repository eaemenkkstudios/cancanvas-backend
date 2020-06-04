package repository

import (
	"context"
	"errors"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepository Interface
type UserRepository interface {
	Save(user *model.NewUser) (*model.User, error)
	FindOne(nickname string) (*model.User, error)
	FindAll() ([]*model.User, error)
	Follow(sender string, target string) (bool, error)
	Unfollow(sender string, target string) (bool, error)
	IsFollowing(sender string, target string) bool
}

type userRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
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

func (db *userRepository) Save(user *model.NewUser) (*model.User, error) {
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
		return nil, errors.New("User " + user.Nickname + " already exists")
	}
	return &model.User{
		Email:    user.Email,
		Artist:   user.Artist,
		Name:     user.Name,
		Nickname: user.Nickname,
	}, nil
}

func (db *userRepository) FindOne(nickname string) (*model.User, error) {
	result := db.collection.FindOne(context.TODO(), bson.M{"_id": nickname})
	var user *UserSchema
	err := result.Decode(&user)
	if err != nil {
		return nil, errors.New("User not found")
	}
	return &model.User{
		Email:    user.Email,
		Name:     user.Name,
		Artist:   user.Artist,
		Nickname: user.Nickname,
	}, nil
}

func (db *userRepository) FindAll() ([]*model.User, error) {
	cursor, err := db.collection.Find(context.TODO(), bson.D{})
	defer cursor.Close(context.TODO())
	var users []*model.User
	for cursor.Next(context.TODO()) {
		var u *model.User
		err = cursor.Decode(&u)
		users = append(users, u)
	}
	return users, err
}

func (db *userRepository) Follow(sender string, target string) (bool, error) {
	if sender == target {
		return false, errors.New("You can't follow yourself")
	}

	result := db.collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var senderUser *UserSchema
	err := result.Decode(&senderUser)
	if err != nil {
		return false, errors.New("Unexpected error")
	}

	result = db.collection.FindOne(context.TODO(), bson.M{"_id": target})
	var targetUser *UserSchema
	err = result.Decode(&targetUser)
	if err != nil {
		return false, errors.New("User " + target + " doesn't exist")
	}

	for _, name := range senderUser.Following {
		if name == target {
			return false, errors.New("You already follow " + target)
		}
	}
	senderUser.Following = append(senderUser.Following, target)
	targetUser.Followers++

	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"following": senderUser.Following},
	})
	if err != nil {
		return false, err
	}

	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": target}, bson.M{
		"$set": bson.M{"followers": targetUser.Followers},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *userRepository) Unfollow(sender string, target string) (bool, error) {
	if sender == target {
		return false, errors.New("You can't unfollow yourself")
	}
	result := db.collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var senderUser *UserSchema
	err := result.Decode(&senderUser)
	if err != nil {
		return false, errors.New("Unexpected error")
	}

	isFollowing := false
	for i, name := range senderUser.Following {
		if name == target {
			isFollowing = true
			senderUser.Following = append(senderUser.Following[:i], senderUser.Following[i+1:]...)
			break
		}
	}

	if !isFollowing {
		return false, errors.New("You don't follow " + target)
	}

	result = db.collection.FindOne(context.TODO(), bson.M{"_id": target})
	var targetUser *UserSchema
	err = result.Decode(&targetUser)
	if err != nil {
		return false, errors.New("User " + target + " doesn't exist")
	}
	targetUser.Followers--

	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"following": senderUser.Following},
	})
	if err != nil {
		return false, err
	}

	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": target}, bson.M{
		"$set": bson.M{"followers": targetUser.Followers},
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *userRepository) IsFollowing(sender string, target string) bool {
	if sender == target {
		return false
	}

	result := db.collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var senderUser *UserSchema
	err := result.Decode(&senderUser)
	if err != nil {
		return false
	}

	for _, name := range senderUser.Following {
		if name == target {
			return true
		}
	}
	return false
}

// NewUserRepository function
func NewUserRepository() UserRepository {
	client := NewDatabaseClient()
	return &userRepository{
		client:     client,
		collection: client.Database(Database).Collection(CollectionUsers),
	}
}
