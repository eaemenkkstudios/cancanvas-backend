package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PostRepository interface
type PostRepository interface {
	GetPosts(sender, author string, page *int) ([]*model.Post, error)
	GetComments(sender, postID string, page *int) ([]*model.PostComment, error)
	CreatePost(author string, content graphql.Upload, description, bidID *string) (string, error)
	EditPost(author, postID, description string) (bool, error)
	DeletePost(author, postID string) (bool, error)
	LikePost(sender, postID string) (bool, error)
	CommentOnPost(sender, postID, message string) (string, error)
	EditComment(sender, postID, commentID, message string) (bool, error)
	DeleteComment(sender, postID, commentID string) (bool, error)
	LikeComment(sender, postID, commentID string) (bool, error)
}

type postRepository struct {
	client     *mongo.Database
	awsSession service.AwsService
}

func (db *postRepository) GetPosts(sender, author string, page *int) ([]*model.Post, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Collection(CollectionPosts)
	opts := options.Find().
		SetSkip(int64(PageSize * (*page - 1))).
		SetLimit(PageSize).
		SetSort(bson.M{"timestamp": -1})
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{"author": author}, opts)
	if err != nil {
		return nil, errors.New("Could not load posts")
	}
	var posts = make([]*model.Post, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var p model.Post
		err = cursor.Decode(&p)
		p.Liked = false
		for _, l := range p.Likes {
			if l == sender {
				p.Liked = true
				break
			}
		}
		posts = append(posts, &p)
	}
	return posts, err
}

func (db *postRepository) GetComments(sender, postID string, page *int) ([]*model.PostComment, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Collection(CollectionPosts)
	id, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, errors.New("Invalid postID")
	}
	result := collection.FindOne(context.TODO(), bson.M{"_id": id})
	var p model.Post
	err = result.Decode(&p)
	if err != nil {
		return nil, errors.New("Post not found")
	}
	commentList := make([]*model.PostComment, 0)
	for i := PageSize * (*page - 1); i < PageSize**page && i < len(p.Comments.List); i++ {
		collection = db.client.Collection(CollectionUsers)
		result = collection.FindOne(context.TODO(), bson.M{"_id": p.Comments.List[i].Author})
		var u UserSchema
		err = result.Decode(&u)
		if err != nil {
			return nil, err
		}
		liked := false
		for _, l := range p.Comments.List[i].Likes {
			if l == sender {
				liked = true
				break
			}
		}
		commentList = append(commentList, &model.PostComment{
			ID: p.Comments.List[i].ID,
			Author: &model.FeedUser{
				Name:     u.Name,
				Nickname: u.Nickname,
				Picture:  u.Picture,
			},
			Likes:     p.Comments.List[i].LikeCount,
			Liked:     liked,
			Text:      p.Comments.List[i].Text,
			Timestamp: p.Comments.List[i].Timestamp,
		})
	}
	return commentList, nil
}

func (db *postRepository) CreatePost(author string, content graphql.Upload, description, bidID *string) (string, error) {
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
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
		BidID:     bidID,
	})
	if err != nil {
		return "", errors.New("Could not create post")
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (db *postRepository) EditPost(author, postID, description string) (bool, error) {
	if description == "" {
		return false, errors.New("Could not edit post")
	}
	collection := db.client.Collection(CollectionPosts)
	id, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return false, errors.New("Invalid postID")
	}
	result, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id, "author": author}, bson.M{
		"$set": bson.M{"description": description},
	})
	if err != nil || result.ModifiedCount == 0 {
		return false, errors.New("Could not edit post")
	}
	return true, nil
}

func (db *postRepository) DeletePost(author, postID string) (bool, error) {
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

func (db *postRepository) LikePost(sender, postID string) (bool, error) {
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

func (db *postRepository) CommentOnPost(sender, postID, message string) (string, error) {
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
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
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

func (db *postRepository) EditComment(sender, postID, commentID, message string) (bool, error) {
	collection := db.client.Collection(CollectionPosts)
	id, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return false, errors.New("Invalid postID")
	}
	result := collection.FindOne(context.TODO(), bson.M{"_id": id})
	var p model.Post
	err = result.Decode(&p)
	if err != nil {
		return false, errors.New("Post not found")
	}
	edited := false
	for i, c := range p.Comments.List {
		if c.ID == commentID && c.Author == sender {
			p.Comments.List[i].Text = message
			p.Comments.List[i].Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
			edited = true
			break
		}
	}
	if !edited {
		return false, errors.New("Comment not found")
	}
	res, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$set": bson.M{"comments": p.Comments},
	})
	if err != nil && res.ModifiedCount > 0 {
		return true, nil
	}
	return false, err
}

func (db *postRepository) DeleteComment(sender, postID, commentID string) (bool, error) {
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

func (db *postRepository) LikeComment(sender, postID, commentID string) (bool, error) {
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
