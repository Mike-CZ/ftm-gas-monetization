package handlers

import (
	"ftm-gas-monetization/internal/config"
	"ftm-gas-monetization/internal/graphql/resolvers"
	gqlSchema "ftm-gas-monetization/internal/graphql/schema"
	"ftm-gas-monetization/internal/logger"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-transport-ws/graphqlws"
	"github.com/rs/cors"
	"net/http"
)

// ApiHandler constructs and return the API HTTP handlers chain for serving GraphQL API calls.
func ApiHandler(cfg *config.Config, log *logger.AppLogger, rs *resolvers.RootResolver) http.Handler {
	// Create new CORS handler and attach the logger into it, so we get information on Debug level if needed
	corsHandler := cors.New(corsOptions(cfg))
	corsHandler.Log = *log

	// we don't want to write a method for each type field if it could be matched directly
	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}

	// create new parsed GraphQL schema
	schema := graphql.MustParseSchema(gqlSchema.Schema(), rs, opts...)

	// return the constructed API handler chain
	return &LoggingHandler{
		logger:  *log,
		handler: corsHandler.Handler(graphqlws.NewHandlerFunc(schema, &relay.Handler{Schema: schema})),
	}
}

// corsOptions constructs new set of options for the CORS handler based on provided configuration.
func corsOptions(cfg *config.Config) cors.Options {
	return cors.Options{
		AllowedOrigins: cfg.Api.CorsOrigin,
		AllowedMethods: []string{"HEAD", "GET", "POST"},
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type", "X-Requested-With"},
		MaxAge:         300,
	}
}
