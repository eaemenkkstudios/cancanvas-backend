package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PostRepository interface
type PostRepository interface {
	CreatePost(author string, content graphql.Upload, description *string) (string, error)
	DeletePost(author string, postID string) (bool, error)
	LikePost(sender string, postID string) (bool, error)
	CommentOnPost(sender string, postID string, message string) (string, error)
	DeleteComment(sender string, postID string, commentID string) (bool, error)
	LikeComment(sender string, postID string, commentID string) (bool, error)
}

type postRepository struct {
	client     *mongo.Database
	awsSession service.AwsService
}

func (db *postRepository) CreatePost(author string, content graphql.Upload, description *string) (string, error) {
	collection := db.client.Collection(CollectionUsers)
	user := collection.FindOne(context.TODO(), bson.M{"_id": author})
	var u UserSchema
	err := user.Decode(&u)
	if err != nil {
		return "", errors.New("Unexpected error")
	}
	filepath, err := db.awsSession.UploadFile(content, "post")
	if err != nil {
		return "", errors.New("Could not upload file")
	}
	collection = db.client.Collection(CollectionPosts)
	result, err := collection.InsertOne(context.TODO(), &model.Post{
		Description: description,
		Author:      author,
		Comments: &model.CommentList{
			List:  make([]*model.Comment, 0),
			Count: 0,
		},
		Content:   filepath,
		LikeCount: 0,
		Likes:     make([]string, 0),
		Timestamp: time.Now(),
	})
	if err != nil {
		return "", errors.New("Could not create post")
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (db *postRepository) DeletePost(author string, postID string) (bool, error) {
	collection := db.client.Collection(CollectionUsers)
	user := collection.FindOne(context.TODO(), bson.M{"_id": author})
	var u UserSchema
	err := user.Decode(&u)
	if err != nil {
		return false, errors.New("Unexpected error")
	}
	collection = db.client.Collection(CollectionPosts)
	id, err := primitive.ObjectIDFromHex(postID)
	result := collection.FindOne(context.TODO(), bson.M{"_id": id, "author": author})
	var post model.Post
	err = result.Decode(&post)
	if err != nil {
		return false, errors.New("Post not found")
	}
	urlPrefix := db.awsSession.GetURLPrefix()
	status, err := db.awsSession.DeleteFile(strings.TrimPrefix(post.Content, urlPrefix))
	if err != nil {
		return status, err
	}
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *postRepository) LikePost(sender string, postID string) (bool, error) {
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return false, errors.New("Unexpected error")
	}
	collection = db.client.Collection(CollectionPosts)
	id, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return false, errors.New("Invalid postID")
	}
	result = collection.FindOne(context.TODO(), bson.M{"_id": id})
	var post model.Post
	err = result.Decode(&post)
	if err != nil {
		return false, errors.New("Post not found")
	}
	likeIndex := -1
	for i, l := range post.Likes {
		if l == sender {
			likeIndex = i
			break
		}
	}
	if likeIndex == -1 {
		post.Likes = append(post.Likes, sender)
		post.LikeCount++
	} else {
		post.Likes = append(post.Likes[:likeIndex], post.Likes[likeIndex+1:]...)
		post.LikeCount--
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$set": bson.M{"likes": post.Likes, "likecount": post.LikeCount},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *postRepository) CommentOnPost(sender string, postID string, message string) (string, error) {
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return "", errors.New("Unexpected error")
	}
	collection = db.client.Collection(CollectionPosts)
	id, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return "", errors.New("Invalid postID")
	}
	result = collection.FindOne(context.TODO(), bson.M{"_id": id})
	var post model.Post
	err = result.Decode(&post)
	if err != nil {
		return "", errors.New("Post not found")
	}
	commentID := primitive.NewObjectID().Hex()
	post.Comments.List = append(post.Comments.List, &model.Comment{
		ID:        commentID,
		Author:    sender,
		Text:      message,
		LikeCount: 0,
		Likes:     make([]string, 0),
		Timestamp: time.Now(),
	})
	post.Comments.Count++
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$set": bson.M{"comments": post.Comments},
	})
	if err != nil {
		return "", err
	}
	return commentID, nil
}

func (db *postRepository) DeleteComment(sender string, postID string, commentID string) (bool, error) {
	collection := db.client.Collection(CollectionUsers)
	user := collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var u UserSchema
	err := user.Decode(&u)
	if err != nil {
		return false, errors.New("Unexpected error")
	}
	collection = db.client.Collection(CollectionPosts)
	id, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return false, errors.New("Invalid postID")
	}
	result := collection.FindOne(context.TODO(), bson.M{"_id": id})
	var post model.Post
	err = result.Decode(&post)
	if err != nil {
		return false, errors.New("Post not found")
	}
	for i, c := range post.Comments.List {
		if c.ID == commentID {
			if c.Author == sender || post.Author == sender {
				post.Comments.List = append(post.Comments.List[:i], post.Comments.List[i+1:]...)
				post.Comments.Count--
			} else {
				return false, errors.New("Unauthorized")
			}
		}
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$set": bson.M{"comments": post.Comments},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *postRepository) LikeComment(sender string, postID string, commentID string) (bool, error) {
	collection := db.client.Collection(CollectionUsers)
	result := collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var u UserSchema
	err := result.Decode(&u)
	if err != nil {
		return false, errors.New("Unexpected error")
	}
	collection = db.client.Collection(CollectionPosts)
	id, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return false, errors.New("Invalid postID")
	}
	result = collection.FindOne(context.TODO(), bson.M{"_id": id})
	var post model.Post
	err = result.Decode(&post)
	if err != nil {
		return false, errors.New("Post not found")
	}
	commentIndex := -1
	likeIndex := -1
	for i, c := range post.Comments.List {
		if c.ID == commentID {
			commentIndex = i
			for index, l := range post.Comments.List[i].Likes {
				if l == sender {
					likeIndex = index
					break
				}
			}
		}
	}
	if commentIndex == -1 {
		return false, errors.New("Comment not found")
	}
	if likeIndex == -1 {
		post.Comments.List[commentIndex].Likes = append(
			post.Comments.List[commentIndex].Likes,
			sender,
		)
		post.Comments.List[commentIndex].LikeCount++
	} else {
		post.Comments.List[commentIndex].Likes = append(
			post.Comments.List[commentIndex].Likes[:likeIndex],
			post.Comments.List[commentIndex].Likes[likeIndex+1:]...,
		)
		post.Comments.List[commentIndex].LikeCount--
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$set": bson.M{"comments.list": post.Comments.List},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// NewPostRepository function
func NewPostRepository() PostRepository {
	client := newDatabaseClient()
	awsSession := service.NewAwsService()
	return &postRepository{
		client,
		awsSession,
	}
}
