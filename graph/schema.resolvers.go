package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/generated"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/repository"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
)

var userRepository = repository.NewUserRepository()
var authRepository = repository.NewAuthRepository()
var jwtService = service.NewJWTService()

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	return userRepository.Save(&input)
}

func (r *mutationResolver) Follow(ctx context.Context, nickname string) (bool, error) {
	token := ctx.Value("token")
	if token == nil {
		return false, errors.New("Unauthorized")
	}
	claims, err := jwtService.GetClaimsFromToken(fmt.Sprintf("%v", token))
	if err != nil {
		return false, errors.New("Unauthorized")
	}
	sender := fmt.Sprintf("%v", claims["name"])
	return userRepository.Follow(sender, nickname)
}

func (r *mutationResolver) Unfollow(ctx context.Context, nickname string) (bool, error) {
	token := ctx.Value("token")
	if token == nil {
		return false, errors.New("Unauthorized")
	}
	claims, err := jwtService.GetClaimsFromToken(fmt.Sprintf("%v", token))
	if err != nil {
		return false, errors.New("Unauthorized")
	}
	sender := fmt.Sprintf("%v", claims["name"])
	return userRepository.Unfollow(sender, nickname)
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	token := ctx.Value("token")
	if token == nil {
		return nil, errors.New("Unauthorized")
	}
	result, err := jwtService.ValidateToken(fmt.Sprintf("%v", token))
	if err != nil || !result.Valid {
		return nil, errors.New("Unauthorized")
	}
	return userRepository.FindAll()
}

func (r *queryResolver) User(ctx context.Context, nickname string) (*model.User, error) {
	token := ctx.Value("token")
	if token == nil {
		return nil, errors.New("Unauthorized")
	}
	result, err := jwtService.ValidateToken(fmt.Sprintf("%v", token))
	if err != nil || !result.Valid {
		return nil, errors.New("Unauthorized")
	}
	return userRepository.FindOne(nickname)
}

func (r *queryResolver) Self(ctx context.Context) (*model.User, error) {
	token := ctx.Value("token")
	if token == nil {
		return nil, errors.New("Unauthorized")
	}
	claims, err := jwtService.GetClaimsFromToken(fmt.Sprintf("%v", token))
	if err != nil {
		return nil, errors.New("Unauthorized")
	}
	nickname := fmt.Sprintf("%v", claims["name"])
	return userRepository.FindOne(nickname)
}

func (r *queryResolver) Login(ctx context.Context, nickname string, password string) (string, error) {
	return authRepository.Login(nickname, password)
}

func (r *queryResolver) IsFollowing(ctx context.Context, nickname string) (bool, error) {
	token := ctx.Value("token")
	if token == nil {
		return false, errors.New("Unauthorized")
	}
	claims, err := jwtService.GetClaimsFromToken(fmt.Sprintf("%v", token))
	if err != nil {
		return false, errors.New("Unauthorized")
	}
	sender := fmt.Sprintf("%v", claims["name"])
	return userRepository.IsFollowing(sender, nickname), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
