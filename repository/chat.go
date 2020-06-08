package repository

import (
	"context"
	"errors"
	"time"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChatRepository interface
type ChatRepository interface {
	SendMessage(sender, msg, receiver string) (bool, error)
	SendMessageToDialogflow(sender, msg string) (string, error)
	NewChatMessage(sender string) (<-chan *model.Message, error)
	createChat(sender, receiver string) (string, error)
	addMessageToChat(chatID, message, sender string) error
	addChatToUser(chatID, sender, receiver string) error
}

type chatRepository struct {
	client            *mongo.Database
	dialogflowService service.DialogflowService
}

// Chat struct
type Chat struct {
	ID       string    `json:"_id,omitempty" bson:"_id,omitempty"`
	Users    []string  `json:"users"`
	Messages []Message `json:"messages"`
}

// Message struct
type Message struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Sender    string    `json:"sender"`
}

type userChat struct {
	ChatID   string `json:"chatid"`
	Receiver string `json:"receiver"`
}

func (db *chatRepository) SendMessage(sender, msg, receiver string) (bool, error) {
	if sender == receiver {
		return false, errors.New("You can't send a message to yourself")
	}
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return false, errors.New("User not found")
	}
	var chat *userChat = nil
	for _, c := range u.Chats {
		if c.Receiver == receiver {
			chat = &c
			break
		}
	}
	if chat == nil {
		result = collection.FindOne(context.TODO(), bson.M{"_id": receiver})
		var r userChat
		err = result.Decode(&r)
		if err != nil {
			return false, errors.New("No user named '" + receiver + "' found")
		}
		chatID, err := db.createChat(sender, receiver)
		if err != nil {
			return false, errors.New("Could not send message")
		}
		err = db.addChatToUser(chatID, sender, receiver)
		if err != nil {
			return false, errors.New("Could not send message")
		}

		err = db.addChatToUser(chatID, receiver, sender)
		if err != nil {
			return false, errors.New("Could not send message")
		}

		err = db.addMessageToChat(chatID, msg, sender)
		if err != nil {
			return false, errors.New("Could not send message")
		}

	} else {
		err = db.addMessageToChat(chat.ChatID, msg, sender)
		if err != nil {
			return false, errors.New("Could not send message")
		}
	}
	return true, nil
}

func (db *chatRepository) SendMessageToDialogflow(sender, msg string) (string, error) {
	result, err := db.dialogflowService.SendMessage(sender, msg)
	return result, err
}

func (db *chatRepository) NewChatMessage(sender string) (<-chan *model.Message, error) {
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var user UserSchema
	err := result.Decode(&user)
	if err != nil {
		return nil, errors.New("Unexpected Error")
	}
	var chatIDs = make([]primitive.ObjectID, 0)
	for _, c := range user.Chats {
		id, _ := primitive.ObjectIDFromHex(c.ChatID)
		chatIDs = append(chatIDs, id)
	}

	collection = db.client.Collection(CollectionChats)
	ctx := context.TODO()
	stream, err := collection.Watch(ctx, mongo.Pipeline{bson.D{
		{
			Key: "$match", Value: bson.D{{Key: "fullDocument._id", Value: bson.D{{Key: "$in", Value: chatIDs}}}},
		},
	}}, options.
		ChangeStream().
		SetMaxAwaitTime(1*time.Hour).
		SetFullDocument(options.FullDocument("updateLookup")))
	if err != nil {
		return nil, errors.New("Unexpected error")
	}
	if stream.Err() != nil {
		return nil, errors.New("Unexpected error")
	}
	messageChan := make(chan *model.Message)
	go func() {
		defer stream.Close(ctx)
		for stream.Next(ctx) {
			var c bson.D
			err = stream.Decode(&c)
			if err != nil {
				messageChan <- nil
				break
			}
			bsonByes, err := bson.Marshal(c.Map()["fullDocument"])
			if err != nil {
				return
			}
			var updatedChat *Chat
			bson.Unmarshal(bsonByes, &updatedChat)
			lastMsg := updatedChat.Messages[len(updatedChat.Messages)-1]
			select {
			case messageChan <- &model.Message{
				ChatID:    updatedChat.ID,
				Message:   lastMsg.Message,
				Sender:    lastMsg.Sender,
				Timestamp: lastMsg.Timestamp,
			}:
			}
		}
	}()
	return messageChan, nil
}

func (db *chatRepository) createChat(sender, receiver string) (string, error) {
	collection := db.client.Collection(CollectionChats)
	result, err := collection.InsertOne(context.TODO(), &Chat{
		Messages: make([]Message, 0),
		Users:    []string{sender, receiver},
	})
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (db *chatRepository) addMessageToChat(chatID, message, sender string) error {
	collection := db.client.Collection(CollectionChats)
	msg := &Message{
		Message:   message,
		Sender:    sender,
		Timestamp: time.Now(),
	}
	id, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return errors.New("Invalid ChatID")
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$push": bson.M{"messages": msg},
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *chatRepository) addChatToUser(chatID, sender, receiver string) error {
	collection := db.client.Collection(CollectionUsers)
	chat := &userChat{
		ChatID:   chatID,
		Receiver: receiver,
	}
	_, err := collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$push": bson.M{"chats": chat},
	})
	if err != nil {
		return err
	}
	return nil
}

// NewChatRepository function
func NewChatRepository() ChatRepository {
	client := newDatabaseClient()
	dialogflowService := service.NewDialogflowService()
	return &chatRepository{
		client,
		dialogflowService,
	}
}
