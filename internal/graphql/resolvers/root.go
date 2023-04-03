// Package resolvers implements GraphQL resolvers to incoming API requests.
package resolvers

import (
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/config"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"sync"
)

// log represents the logger to be used by the repository.
var log logger.AppLogger

// config represents the configuration setup used by the repository
// to establish and maintain required connectivity to external services
// as needed.
var cfg *config.Config

// resolver represents a singleton resolver of the root resolver
var resolver RootResolver

// oneInstance is the sync guarding root resolver singleton creation.
var oneInstance sync.Once

type RootResolver struct{}

// SetLogger sets the repository logger to be used to collect logging info.
func SetLogger(l logger.AppLogger) {
	log = *l.ModuleLogger("graphql")
}

// SetConfig sets the repository configuration to be used to establish
// and maintain external repository connections.
func SetConfig(c *config.Config) {
	cfg = c
}

// Resolver returns a singleton resolver fo the root resolver.
func Resolver() *RootResolver {
	// make sure to instantiate the Repository only once
	oneInstance.Do(func() {
		resolver = *newResolver()
	})
	return &resolver
}

// new creates a resolver of root resolver and initializes its internal structure.
func newResolver() *RootResolver {
	if cfg == nil {
		panic(fmt.Errorf("missing configuration"))
	}
	if &log == nil {
		panic(fmt.Errorf("missing logger"))
	}

	// create new resolver
	rs := RootResolver{}
	log.Notice("GraphQL resolver started")

	return &rs
}
