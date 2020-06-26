package repository

import (
	"context"
	"errors"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TagsRepository interface
type TagsRepository interface {
	GetTags() ([]string, error)
	GetUserTags(nickname string) ([]string, error)
	GetUsersPerTags(tags []string, page *int) ([]*model.User, error)
	UpdateUserTags(user string, tags []string) (bool, error)
	AddTagToUser(user, tag string) (bool, error)
	RemoveTagFromUser(user, tag string) (bool, error)
}

type tagsRepository struct {
	client *mongo.Database
}

// TagSchema struct
type TagSchema struct {
	ID    string   `json:"_id" bson:"_id"`
	Users []string `json:"users"`
}

func (db *tagsRepository) GetTags() ([]string, error) {
	collection := db.client.Collection(CollectionTags)
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.New("Could not get tags")
	}
	tags := make([]string, 0)
	var t TagSchema
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		err = cursor.Decode(&t)
		tags = append(tags, t.ID)
	}
	return tags, err
}

func (db *tagsRepository) GetUserTags(nickname string) ([]string, error) {
	collection := db.client.Collection(CollectionTags)
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{"users": bson.M{"$in": []string{nickname}}})
	if err != nil {
		return nil, errors.New("Could not get user tags")
	}
	defer cursor.Close(ctx)
	tags := make([]string, 0)
	for cursor.Next(ctx) {
		var t TagSchema
		err = cursor.Decode(&t)
		tags = append(tags, t.ID)
	}
	return tags, err
}

func (db *tagsRepository) GetUsersPerTags(tags []string, page *int) ([]*model.User, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Collection(CollectionTags)
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": tags}})
	if err != nil {
		return nil, err
	}
	tagList := make([]*TagSchema, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var t TagSchema
		err = cursor.Decode(&t)
		tagList = append(tagList, &t)
	}
	if err != nil {
		return nil, err
	}
	users := make([]string, 0)
	for _, tg := range tagList {
		for _, u := range tg.Users {
			if isInList := stringInSlice(u, users); !isInList {
				users = append(users, u)
			}
		}
	}

	collection = db.client.Collection(CollectionUsers)
	opts := options.Find().
		SetSkip(int64(PageSize * (*page - 1))).
		SetLimit(PageSize).
		SetSort(bson.M{"followerscount": -1})
	usersCtx := context.TODO()
	usersCursor, err := collection.Find(usersCtx, bson.M{"_id": bson.M{"$in": users}}, opts)
	if err != nil {
		return nil, err
	}
	defer usersCursor.Close(usersCtx)
	userList := make([]*model.User, 0)
	for usersCursor.Next(usersCtx) {
		var user UserSchema
		err = usersCursor.Decode(&user)
		userList = append(userList, &model.User{
			Nickname:       user.Nickname,
			Name:           user.Name,
			Bio:            user.Bio,
			Email:          user.Email,
			Cover:          user.Cover,
			Picture:        user.Picture,
			Followers:      user.Followers,
			FollowersCount: user.FollowersCount,
			Following:      user.Following,
			Lat:            user.Lat,
			Lng:            user.Lng,
		})
	}
	return userList, nil
}

func (db *tagsRepository) UpdateUserTags(user string, tags []string) (bool, error) {
	collection := db.client.Collection(CollectionTags)
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{"users": bson.M{"$in": []string{user}}})
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var t TagSchema
		err = cursor.Decode(&t)
		for i, u := range t.Users {
			if u == user {
				t.Users = append(t.Users[:i], t.Users[i+1:]...)
				break
			}
		}
		collection.UpdateOne(ctx, bson.M{"_id": t.ID}, bson.M{
			"$set": bson.M{"users": t.Users},
		})
	}
	if err != nil {
		return false, err
	}
	cursor, err = collection.Find(ctx, bson.M{"_id": bson.M{"$in": tags}})
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var t TagSchema
		err = cursor.Decode(&t)
		t.Users = append(t.Users, user)
		collection.UpdateOne(ctx, bson.M{"_id": t.ID}, bson.M{
			"$set": bson.M{"users": t.Users},
		})
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *tagsRepository) AddTagToUser(user, tag string) (bool, error) {
	collection := db.client.Collection(CollectionTags)
	result := collection.FindOne(context.TODO(), bson.M{"_id": tag})
	var t TagSchema
	err := result.Decode(&t)
	if err != nil {
		return false, errors.New("Tag not found")
	}
	for _, u := range t.Users {
		if u == user {
			return false, errors.New("You already have this tag")
		}
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": tag}, bson.M{
		"$push": bson.M{"users": user},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *tagsRepository) RemoveTagFromUser(user, tag string) (bool, error) {
	collection := db.client.Collection(CollectionTags)
	result := collection.FindOne(context.TODO(), bson.M{"_id": tag})
	var t TagSchema
	err := result.Decode(&t)
	if err != nil {
		return false, errors.New("Tag not found")
	}
	for i, u := range t.Users {
		if u == user {
			t.Users = append(t.Users[:i], t.Users[i+1:]...)
		}
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": tag}, bson.M{
		"$set": bson.M{"users": t.Users},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// NewTagsRepository function
func NewTagsRepository() TagsRepository {
	client := newDatabaseClient()
	return &tagsRepository{
		client,
	}
}
