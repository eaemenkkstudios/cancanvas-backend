package repository

import (
	"context"
	"errors"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// FeedRepository interface
type FeedRepository interface {
	GetFeed(nickname string, page *int) ([]*model.FeedPost, error)
	GetTrending(nickname string, page *int) ([]*model.FeedPost, error)
}

type feedRepository struct {
	client *mongo.Database
}

// PageSize constant
const PageSize = 10

type feedPost struct {
	ID          string             `bson:"_id"`
	Author      []*UserSchema      `bson:"author"`
	Description *string            `bson:"description"`
	Content     string             `bson:"content"`
	Timestamp   string             `bson:"timestamp"`
	Comments    *model.CommentList `bson:"comments"`
	LikeCount   int                `bson:"likecount"`
	Likes       []string           `bson:"likes"`
	BidID       *string            `bson:"bidID"`
}

func (db *feedRepository) GetFeed(nickname string, page *int) ([]*model.FeedPost, error) {
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
	ctx := context.TODO()
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"author": bson.M{"$in": u.Following}}}},
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         CollectionUsers,
				"localField":   "author",
				"foreignField": "_id",
				"as":           "author",
			}}},
		bson.D{{Key: "$skip", Value: int64(PageSize * (*page - 1))}},
		bson.D{{Key: "$limit", Value: PageSize}},
		bson.D{{Key: "$sort", Value: bson.M{"timestamp": -1}}},
	})
	if err != nil {
		return nil, errors.New("Could not load feed")
	}
	var posts = make([]*model.FeedPost, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var p feedPost
		err = cursor.Decode(&p)
		liked := false
		for _, l := range p.Likes {
			if l == nickname {
				liked = true
				break
			}
		}
		posts = append(posts, &model.FeedPost{
			ID: p.ID,
			Author: &model.FeedUser{
				Name:     p.Author[0].Name,
				Nickname: p.Author[0].Nickname,
				Picture:  p.Author[0].Picture,
			},
			Comments:    p.Comments,
			Content:     p.Content,
			Description: p.Description,
			Likes:       p.LikeCount,
			Liked:       liked,
			Timestamp:   p.Timestamp,
			BidID:       p.BidID,
		})
	}
	return posts, err
}

func (db *feedRepository) GetTrending(nickname string, page *int) ([]*model.FeedPost, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Collection(CollectionPosts)
	ctx := context.TODO()
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         CollectionUsers,
				"localField":   "author",
				"foreignField": "_id",
				"as":           "author",
			}}},
		bson.D{{Key: "$skip", Value: int64(PageSize * (*page - 1))}},
		bson.D{{Key: "$limit", Value: PageSize}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "likes", Value: -1}, {Key: "timestamp", Value: -1}}}},
	})
	if err != nil {
		return nil, errors.New("Could not load feed")
	}
	var posts = make([]*model.FeedPost, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var p feedPost
		err = cursor.Decode(&p)
		liked := false
		for _, l := range p.Likes {
			if l == nickname {
				liked = true
				break
			}
		}
		posts = append(posts, &model.FeedPost{
			ID: p.ID,
			Author: &model.FeedUser{
				Name:     p.Author[0].Name,
				Nickname: p.Author[0].Nickname,
				Picture:  p.Author[0].Picture,
			},
			Comments:    p.Comments,
			Content:     p.Content,
			Description: p.Description,
			Likes:       p.LikeCount,
			Liked:       liked,
			Timestamp:   p.Timestamp,
			BidID:       p.BidID,
		})
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
