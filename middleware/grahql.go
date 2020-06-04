package middleware

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/eaemenkkstudios/cancanvas-backend/graph"
	"github.com/eaemenkkstudios/cancanvas-backend/graph/generated"
	"github.com/gin-gonic/gin"
)

// PlaygroundHandler function
func PlaygroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL playground", "/query")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// GraphQLHandler function
func GraphQLHandler() gin.HandlerFunc {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		token = strings.TrimSpace(token)
		if token != "" {
			ctx := context.WithValue(c.Request.Context(), "token", token)
			c.Request = c.Request.WithContext(ctx)
		}
		srv.ServeHTTP(c.Writer, c.Request)
	}
}
