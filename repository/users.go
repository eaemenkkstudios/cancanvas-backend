package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository Interface
type UserRepository interface {
	CreateUser(user *model.NewUser) (*model.User, error)
	FindOne(nickname string) (*model.User, error)
	FindAll(nickname *string, page *int) ([]*model.User, error)
	Follow(sender, target string) (bool, error)
	Unfollow(sender, target string) (bool, error)
	IsFollowing(sender, target string) bool
	UpdateProfilePicture(sender string, picture graphql.Upload) (string, error)
	UpdateCover(sender string, cover graphql.Upload) (string, error)
	UpdateLocation(sender string, lat, lng float64) (bool, error)
	UpdateBio(sender, bio string) (bool, error)
}

type userRepository struct {
	client     *mongo.Database
	awsSession service.AwsService
	collection *mongo.Collection
}

// UserSchema struct
type UserSchema struct {
	Nickname       string     `json:"nickname" bson:"_id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Picture        string     `json:"picture"`
	Cover          string     `json:"cover"`
	Bio            string     `json:"bio"`
	Followers      []string   `json:"followers"`
	FollowersCount int        `json:"followerscount"`
	Following      []string   `json:"following"`
	Password       Password   `json:"password"`
	Chats          []userChat `json:"chats"`
	First          bool       `json:"first"`
	Lat            float64    `json:"lat"`
	Lng            float64    `json:"lng"`
}

func (db *userRepository) CreateUser(user *model.NewUser) (*model.User, error) {
	u := &UserSchema{
		Email:          user.Email,
		Nickname:       strings.ToLower(user.Nickname),
		Name:           user.Name,
		Following:      make([]string, 0),
		FollowersCount: 0,
		Followers:      make([]string, 0),
		Chats:          make([]userChat, 0),
		First:          true,
		Picture:        "",
		Cover:          "",
		Bio:            "",
		Lat:            0,
		Lng:            0,
		Password:       GeneratePassword(user.Password),
	}
	_, err := db.collection.InsertOne(context.TODO(), u)
	if err != nil {
		return nil, errors.New("User '" + user.Nickname + "' already exists")
	}
	return &model.User{
		Email:          user.Email,
		Name:           user.Name,
		Nickname:       user.Nickname,
		Followers:      u.Followers,
		FollowersCount: u.FollowersCount,
		Following:      u.Following,
		Picture:        u.Picture,
		Cover:          u.Cover,
		Bio:            u.Bio,
		Lat:            u.Lat,
		Lng:            u.Lng,
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
		Email:          user.Email,
		Name:           user.Name,
		Nickname:       user.Nickname,
		Followers:      user.Followers,
		FollowersCount: user.FollowersCount,
		Following:      user.Following,
		Picture:        user.Picture,
		Cover:          user.Cover,
		Bio:            user.Bio,
		Lat:            user.Lat,
		Lng:            user.Lng,
	}, nil
}

func (db *userRepository) FindAll(nickname *string, page *int) ([]*model.User, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	document := bson.M{}
	if nickname != nil && *nickname != "" {
		document = bson.M{"_id": bson.M{"$regex": nickname}}
	}
	opts := options.Find().
		SetSkip(int64(PageSize * (*page - 1))).
		SetLimit(PageSize)
	ctx := context.TODO()
	cursor, err := db.collection.Find(ctx, document, opts)
	defer cursor.Close(ctx)
	var users []*model.User
	for cursor.Next(ctx) {
		var u *UserSchema
		err = cursor.Decode(&u)
		users = append(users, &model.User{
			Nickname:       u.Nickname,
			Name:           u.Name,
			Email:          u.Email,
			Followers:      u.Followers,
			FollowersCount: u.FollowersCount,
			Following:      u.Following,
			Picture:        u.Picture,
			Cover:          u.Cover,
			Bio:            u.Bio,
			Lat:            u.Lat,
			Lng:            u.Lng,
		})
	}
	return users, err
}

func (db *userRepository) Follow(sender, target string) (bool, error) {
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
		return false, errors.New("No user name '" + target + "' found")
	}

	for _, name := range senderUser.Following {
		if name == target {
			return false, errors.New("You already follow '" + target + "'")
		}
	}
	senderUser.Following = append(senderUser.Following, target)
	targetUser.Followers = append(targetUser.Followers, sender)
	targetUser.FollowersCount++

	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"following": senderUser.Following},
	})
	if err != nil {
		return false, err
	}

	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": target}, bson.M{
		"$set": bson.M{
			"followers":      targetUser.Followers,
			"followerscount": targetUser.FollowersCount,
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *userRepository) Unfollow(sender, target string) (bool, error) {
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
		return false, errors.New("You don't follow '" + target + "'")
	}

	result = db.collection.FindOne(context.TODO(), bson.M{"_id": target})
	var targetUser *UserSchema
	err = result.Decode(&targetUser)
	if err != nil {
		return false, errors.New("No user named '" + target + "' found")
	}
	targetUser.FollowersCount--
	for i, name := range targetUser.Followers {
		if name == sender {
			targetUser.Followers = append(targetUser.Followers[:i], targetUser.Followers[i+1:]...)
		}
	}
	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"following": senderUser.Following},
	})
	if err != nil {
		return false, err
	}

	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": target}, bson.M{
		"$set": bson.M{"followerscount": targetUser.FollowersCount, "followers": targetUser.Followers},
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *userRepository) IsFollowing(sender, target string) bool {
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

func (db *userRepository) UpdateProfilePicture(sender string, picture graphql.Upload) (string, error) {
	result := db.collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var user UserSchema
	err := result.Decode(&user)
	if err != nil {
		return "", errors.New("User not found")
	}
	newPictureURL, err := db.awsSession.UploadFile(picture, "profile")
	if err != nil {
		return "", err
	}
	urlPrefix := db.awsSession.GetURLPrefix()
	if user.Picture != "" {
		_, err := db.awsSession.DeleteFile(strings.TrimPrefix(user.Picture, urlPrefix))
		if err != nil {
			return "", err
		}
	}
	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"picture": newPictureURL},
	})
	if err != nil {
		return "", err
	}
	return newPictureURL, nil
}

func (db *userRepository) UpdateLocation(sender string, lat, lng float64) (bool, error) {
	_, err := db.collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"lat": lat, "lng": lng},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *userRepository) UpdateBio(sender, bio string) (bool, error) {
	_, err := db.collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"bio": bio},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *userRepository) UpdateCover(sender string, cover graphql.Upload) (string, error) {
	result := db.collection.FindOne(context.TODO(), bson.M{"_id": sender})
	var user UserSchema
	err := result.Decode(&user)
	if err != nil {
		return "", errors.New("User not found")
	}
	newPictureURL, err := db.awsSession.UploadFile(cover, "cover")
	if err != nil {
		return "", err
	}
	urlPrefix := db.awsSession.GetURLPrefix()
	if user.Cover != "" {
		_, err := db.awsSession.DeleteFile(strings.TrimPrefix(user.Cover, urlPrefix))
		if err != nil {
			return "", err
		}
	}
	_, err = db.collection.UpdateOne(context.TODO(), bson.M{"_id": sender}, bson.M{
		"$set": bson.M{"cover": newPictureURL},
	})
	if err != nil {
		return "", err
	}
	return newPictureURL, nil
}

// NewUserRepository function
func NewUserRepository() UserRepository {
	client := newDatabaseClient()
	awsSession := service.NewAwsService()
	return &userRepository{
		client:     client,
		awsSession: awsSession,
		collection: client.Collection(CollectionUsers),
	}
}
