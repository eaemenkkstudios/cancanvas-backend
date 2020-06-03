package middleware

import (
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
		srv.ServeHTTP(c.Writer, c.Request)
	}
}
