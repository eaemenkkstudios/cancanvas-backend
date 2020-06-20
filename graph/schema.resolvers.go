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
var tagsRepository = repository.NewTagsRepository()
var auctionRepository = repository.NewAuctionRepository()

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	return userRepository.CreateUser(&input)
}

func (r *mutationResolver) UpdateUserPicture(ctx context.Context, picture graphql.Upload) (string, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return "", err
	}
	return userRepository.UpdateProfilePicture(sender, picture)
}

func (r *mutationResolver) UpdateUserLocation(ctx context.Context, lat float64, lng float64) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return userRepository.UpdateLocation(sender, lat, lng)
}

func (r *mutationResolver) UpdateUserBio(ctx context.Context, bio string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return userRepository.UpdateBio(sender, bio)
}

func (r *mutationResolver) UpdateUserCover(ctx context.Context, cover graphql.Upload) (string, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return "", err
	}
	return userRepository.UpdateCover(sender, cover)
}

func (r *mutationResolver) AddTagToUser(ctx context.Context, tag string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return tagsRepository.AddTagToUser(sender, tag)
}

func (r *mutationResolver) RemoveTagFromUser(ctx context.Context, tag string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return tagsRepository.RemoveTagFromUser(sender, tag)
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

func (r *mutationResolver) SendMessageToDialogflow(ctx context.Context, msg string) (string, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return "", err
	}
	return chatRepository.SendMessageToDialogflow(sender, msg)
}

func (r *mutationResolver) CreatePost(ctx context.Context, content graphql.Upload, description *string) (string, error) {
	author, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return "", err
	}
	return postRepository.CreatePost(author, content, description)
}

func (r *mutationResolver) EditPost(ctx context.Context, postID string, description string) (bool, error) {
	author, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return postRepository.EditPost(author, postID, description)
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

func (r *mutationResolver) CreateAuction(ctx context.Context, offer float64, description string) (*model.Auction, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return auctionRepository.CreateAuction(sender, description, offer)
}

func (r *mutationResolver) DeleteAuction(ctx context.Context, auctionID string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return auctionRepository.DeleteAuction(sender, auctionID)
}

func (r *mutationResolver) CreateBid(ctx context.Context, auctionID string, deadline string, price float64) (*model.Bid, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return auctionRepository.CreateBid(sender, auctionID, deadline, price)
}

func (r *mutationResolver) DeleteBid(ctx context.Context, auctionID string, bidID string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return auctionRepository.DeleteBid(sender, auctionID, bidID)
}

func (r *mutationResolver) AcceptBid(ctx context.Context, auctionID string, bidID string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return auctionRepository.AcceptBid(sender, auctionID, bidID)
}

func (r *mutationResolver) SendForgotPasswordEmail(ctx context.Context, nickname string) (bool, error) {
	return authRepository.SendForgotPasswordEmail(nickname)
}

func (r *mutationResolver) ResetPassword(ctx context.Context, token string, newPassword string) (bool, error) {
	sender, hash, err := utils.GetSenderAndHashFromToken(token)
	if err != nil {
		return false, err
	}
	return authRepository.ResetPassword(sender, hash, newPassword)
}

func (r *queryResolver) Users(ctx context.Context, nickname *string, page *int) ([]*model.User, error) {
	_, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return userRepository.FindAll(nickname, page)
}

func (r *queryResolver) Self(ctx context.Context) (*model.User, error) {
	nickname, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return userRepository.FindOne(nickname)
}

func (r *queryResolver) Feed(ctx context.Context, page *int) ([]*model.FeedPost, error) {
	nickname, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return feedRepository.GetFeed(nickname, page)
}

func (r *queryResolver) Trending(ctx context.Context, page *int) ([]*model.FeedPost, error) {
	return feedRepository.GetTrending(page)
}

func (r *queryResolver) User(ctx context.Context, nickname string) (*model.User, error) {
	nickname, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return userRepository.FindOne(nickname)
}

func (r *queryResolver) UserPosts(ctx context.Context, nickname string, page *int) ([]*model.Post, error) {
	return postRepository.GetPosts(nickname, page)
}

func (r *queryResolver) Tags(ctx context.Context) ([]string, error) {
	return tagsRepository.GetTags()
}

func (r *queryResolver) UsersByTags(ctx context.Context, tags []string, page *int) ([]*model.User, error) {
	return tagsRepository.GetUsersPerTags(tags, page)
}

func (r *queryResolver) Auctions(ctx context.Context, page *int) ([]*model.FeedAuction, error) {
	return auctionRepository.GetAuctions(page)
}

func (r *queryResolver) Login(ctx context.Context, nickname string, password string) (*model.Login, error) {
	return authRepository.Login(nickname, password)
}

func (r *queryResolver) IsFollowing(ctx context.Context, nickname string) (bool, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return false, err
	}
	return userRepository.IsFollowing(sender, nickname), nil
}

func (r *queryResolver) AcceptedBids(ctx context.Context) ([]*model.FeedAuction, error) {
	sender, err := utils.GetSenderFromTokenHTTP(ctx)
	if err != nil {
		return nil, err
	}
	return auctionRepository.AcceptedBids(sender)
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
