package repository

import (
	"context"
	"errors"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UploadRepository interface
type UploadRepository interface {
	CreatePost(author string, content graphql.Upload, description *string) (string, error)
}

type uploadRepository struct {
	client     *mongo.Client
	awsSession service.AwsService
}

func (db *uploadRepository) CreatePost(author string, content graphql.Upload, description *string) (string, error) {
	filepath, err := db.awsSession.UploadFile(content)
	if err != nil {
		return "", errors.New("Could not upload file")
	}
	collection := db.client.Database(Database).Collection(CollectionUsers)
	user := collection.FindOne(context.TODO(), bson.M{"_id": author})
	var u UserSchema
	err = user.Decode(&u)
	if err != nil {
		return "", errors.New("Unexpected error")
	}
	collection = db.client.Database(Database).Collection(CollectionPosts)
	result, err := collection.InsertOne(context.TODO(), &model.Post{
		Description: description,
		Author:      author,
		Comments: &model.CommentList{
			List:  make([]*model.Comment, 0),
			Count: 0,
		},
		Content: filepath,
		Reactions: &model.ReactionList{
			List:  make([]*model.Reaction, 0),
			Count: make([]*model.ReactionCount, 0),
		},
		Timestamp: time.Now(),
	})
	if err != nil {
		return "", errors.New("Could not create post")
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// NewUploadRepository function
func NewUploadRepository() UploadRepository {
	client := newDatabaseClient()
	awsSession := service.NewAwsService()
	return &uploadRepository{
		client,
		awsSession,
	}
}
