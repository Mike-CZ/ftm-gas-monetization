package api

import (
	"ftm-gas-monetization/internal/config"
	"ftm-gas-monetization/internal/graphql/resolvers"
	"ftm-gas-monetization/internal/handlers"
	"ftm-gas-monetization/internal/logger"
	"ftm-gas-monetization/internal/repository"
	"net/http"
	"time"
)

type ApiServer struct {
	cfg    *config.Config
	log    *logger.AppLogger
	srv    *http.Server
	api    *resolvers.RootResolver
	closed chan interface{}
}

func New(cfg *config.Config, log *logger.AppLogger) *ApiServer {
	api := new(ApiServer)
	api.cfg = cfg
	api.log = log

	resolvers.SetConfig(cfg)
	resolvers.SetLogger(*log)

	repository.SetConfig(cfg)
	repository.SetLogger(*log)

	api.api = resolvers.Resolver()
	api.init()
	return api
}

func (api *ApiServer) init() {
	repository.New(api.cfg, api.log)

	api.closed = make(chan interface{})
	api.makeHttpServer()
	api.run()

}

// makeHttpServer creates and configures the HTTP server to be used to serve incoming requests
func (api *ApiServer) makeHttpServer() {
	// create request MUXer
	srvMux := http.NewServeMux()

	h := http.TimeoutHandler(
		handlers.ApiHandler(api.cfg, api.log, api.api),
		time.Second*time.Duration(api.cfg.Api.ResolverTimeout),
		"Service timeout.",
	)

	srvMux.Handle("/api", h)
	srvMux.Handle("/graphql", h)

	// handle GraphiQL interface
	srvMux.Handle("/graphi", handlers.GraphiHandler(api.cfg.Api.DomainAddress, *api.log))

	// create HTTP server to handle our requests
	api.srv = &http.Server{
		Addr:              api.cfg.Api.BindAddress,
		ReadTimeout:       time.Second * time.Duration(api.cfg.Api.ReadTimeout),
		WriteTimeout:      time.Second * time.Duration(api.cfg.Api.WriteTimeout),
		IdleTimeout:       time.Second * time.Duration(api.cfg.Api.IdleTimeout),
		ReadHeaderTimeout: time.Second * time.Duration(api.cfg.Api.HeaderTimeout),
		Handler:           srvMux,
	}

}

// run executes the API server function.
func (api *ApiServer) run() {
	err := api.srv.ListenAndServe()
	if err != nil {
		api.log.Error(err.Error())
	}
}
