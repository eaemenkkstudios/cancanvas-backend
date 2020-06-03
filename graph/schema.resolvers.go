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
	user := userRepository.Save(&input)
	return user, nil
}

func (r *mutationResolver) Login(ctx context.Context, input *model.Login) (string, error) {
	return authRepository.Login(input.Nickname, input.Password)
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	return userRepository.FindAll(), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
