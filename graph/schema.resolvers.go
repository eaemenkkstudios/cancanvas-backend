package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/generated"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/repository"
	"github.com/eaemenkkstudios/cancanvas-backend/utils"
)

var userRepository = repository.NewUserRepository()
var authRepository = repository.NewAuthRepository()
var chatRepository = repository.NewChatRepository()
var postRepository = repository.NewPostRepository()
var feedRepository = repository.NewFeedRepository()

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	return userRepository.Save(&input)
}

func (r *mutationResolver) Follow(ctx context.Context, nickname string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return userRepository.Follow(sender, nickname)
}

func (r *mutationResolver) Unfollow(ctx context.Context, nickname string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return userRepository.Unfollow(sender, nickname)
}

func (r *mutationResolver) SendMessage(ctx context.Context, msg string, receiver string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return chatRepository.SendMessage(sender, msg, receiver)
}

func (r *mutationResolver) CreatePost(ctx context.Context, content graphql.Upload, description *string) (string, error) {
	author, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return "", err
	}
	return postRepository.CreatePost(author, content, description)
}

func (r *mutationResolver) DeletePost(ctx context.Context, postID string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return postRepository.DeletePost(sender, postID)
}

func (r *mutationResolver) LikeComment(ctx context.Context, postID string, commentID string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return postRepository.LikeComment(sender, postID, commentID)
}

func (r *mutationResolver) LikePost(ctx context.Context, postID string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return postRepository.LikePost(sender, postID)
}

func (r *mutationResolver) CommentOnPost(ctx context.Context, postID string, message string) (string, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return "", err
	}
	return postRepository.CommentOnPost(sender, postID, message)
}

func (r *mutationResolver) DeleteComment(ctx context.Context, postID string, commentID string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return postRepository.DeleteComment(sender, postID, commentID)
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	_, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return userRepository.FindAll()
}

func (r *queryResolver) Self(ctx context.Context) (*model.User, error) {
	nickname, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return userRepository.FindOne(nickname)
}

func (r *queryResolver) Feed(ctx context.Context, page *int) ([]*model.Post, error) {
	nickname, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return feedRepository.GetFeed(nickname, page)
}

func (r *queryResolver) Trending(ctx context.Context, page *int) ([]*model.Post, error) {
	return feedRepository.GetTrending(page)
}

func (r *queryResolver) User(ctx context.Context, nickname string) (*model.User, error) {
	nickname, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return userRepository.FindOne(nickname)
}

func (r *queryResolver) Login(ctx context.Context, nickname string, password string) (string, error) {
	return authRepository.Login(nickname, password)
}

func (r *queryResolver) IsFollowing(ctx context.Context, nickname string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return userRepository.IsFollowing(sender, nickname), nil
}

func (r *subscriptionResolver) NewChatMessage(ctx context.Context) (<-chan *model.Message, error) {
	sender, err := utils.GetSenderFromTokenSocket(ctx)
	if err != nil {
		return nil, err
	}
	return chatRepository.NewChatMessage(sender)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
