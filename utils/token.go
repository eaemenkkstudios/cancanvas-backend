package utils

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/handler"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
)

var jwtService = service.NewJWTService()

// GetSenderFromTokenHTTP function
func GetSenderFromTokenHTTP(ctx context.Context) (string, error) {
	token := ctx.Value("token")
	if fmt.Sprintf("%v", token) == "<nil>" {
		return "", errors.New("Unauthorized")
	}
	return getSenderFromClaims(fmt.Sprintf("%v", token))
}

// GetSenderFromTokenSocket function
func GetSenderFromTokenSocket(ctx context.Context) (string, error) {
	token := handler.GetInitPayload(ctx).GetString("Authorization")
	if token == "" {
		return "", errors.New("Unauthorized")
	}
	return getSenderFromClaims(token)
}

// GetSenderAndHashFromToken function
func GetSenderAndHashFromToken(token string) (sender string, hash string, err error) {
	claims, err := jwtService.GetClaimsFromToken(token)
	if err != nil {
		return "", "", errors.New("Invalid or expired token")
	}
	return fmt.Sprintf("%v", claims["name"]), fmt.Sprintf("%v", claims["hash"]), nil
}

func getSenderFromClaims(token string) (string, error) {
	claims, err := jwtService.GetClaimsFromToken(token)
	if err != nil {
		return "", errors.New("Unauthorized")
	}
	return fmt.Sprintf("%v", claims["name"]), nil
}
