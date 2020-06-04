package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/generated"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/repository"
)

var userRepository repository.UserRepository = repository.NewUserRepository()
var authRepository repository.AuthRepository = repository.NewAuthRepository()

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	return userRepository.Save(&input)
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	return userRepository.FindAll()
}

func (r *queryResolver) User(ctx context.Context, nickname string) (*model.User, error) {
	return userRepository.FindOne(nickname)
}

func (r *queryResolver) Login(ctx context.Context, nickname string, password string) (string, error) {
	return authRepository.Login(nickname, password)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
