package repository

import (
	"context"
	"errors"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FeedRepository interface
type FeedRepository interface {
	GetFeed(nickname string, page *int) ([]*model.Post, error)
}

type feedRepository struct {
	client *mongo.Client
}

// PageSize constant
const PageSize = 10

func (db *feedRepository) GetFeed(nickname string, page *int) ([]*model.Post, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Database(Database).Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": nickname})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return nil, errors.New("Unexpected Error")
	}
	collection = db.client.Database(Database).Collection(CollectionPosts)
	opts := options.Find().SetSkip(int64(PageSize * (*page - 1))).SetLimit(PageSize)
	cursor, err := collection.Find(context.TODO(), bson.M{"author": bson.M{"$in": u.Following}}, opts)
	if err != nil {
		return nil, errors.New("Could not load feed")
	}
	var posts = make([]*model.Post, 0)
	for cursor.Next(context.TODO()) {
		var p model.Post
		err = cursor.Decode(&p)
		posts = append(posts, &p)
	}
	return posts, err
}

// NewFeedRepository function
func NewFeedRepository() FeedRepository {
	client := newDatabaseClient()
	return &feedRepository{
		client,
	}
}