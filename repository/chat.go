package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatRepository interface {
	SendMessage(sender string, msg string, receiver string) (bool, error)
}

type chatRepository struct {
	client *mongo.Client
}

func (db *chatRepository) SendMessage(sender string, msg string, receiver string) (bool, error) {
	collection := db.client.Database(Database).Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return false, errors.New("User not found")
	}
	/* for i, c := range u.Chats {

	} */
	return true, nil
}

func NewChatRepository() ChatRepository {
	client := NewDatabaseClient()
	return &chatRepository{
		client,
	}
}
