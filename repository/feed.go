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
	GetTrending(page *int) ([]*model.Post, error)
}

type feedRepository struct {
	client *mongo.Database
}

// PageSize constant
const PageSize = 10

func (db *feedRepository) GetFeed(nickname string, page *int) ([]*model.Post, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": nickname})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return nil, errors.New("Unexpected Error")
	}
	collection = db.client.Collection(CollectionPosts)
	opts := options.Find().
		SetSkip(int64(PageSize * (*page - 1))).
		SetLimit(PageSize).
		SetSort(bson.M{"timestamp": -1})
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{"author": bson.M{"$in": u.Following}}, opts)
	if err != nil {
		return nil, errors.New("Could not load feed")
	}
	var posts = make([]*model.Post, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var p model.Post
		err = cursor.Decode(&p)
		posts = append(posts, &p)
	}
	return posts, err
}

func (db *feedRepository) GetTrending(page *int) ([]*model.Post, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Collection(CollectionPosts)
	opts := options.Find().
		SetSkip(int64(PageSize * (*page - 1))).
		SetLimit(PageSize).
		SetSort(bson.D{{Key: "likes", Value: -1}, {Key: "timestamp", Value: -1}})
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, errors.New("Could not load feed")
	}
	var posts = make([]*model.Post, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
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
