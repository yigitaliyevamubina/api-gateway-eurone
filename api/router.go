package api

import (
	"fourth-exam/api_gateway_evrone/api/handlers"
	v1 "fourth-exam/api_gateway_evrone/api/handlers/v1"
	"fourth-exam/api_gateway_evrone/api/middleware"
	grpserviceclient "fourth-exam/api_gateway_evrone/internal/infrastructure/grp_service_client"
	"fourth-exam/api_gateway_evrone/internal/infrastructure/repository/redis"
	"fourth-exam/api_gateway_evrone/internal/pkg/config"
	"fourth-exam/api_gateway_evrone/internal/usecase/event"
	"net/http"
	"time"

	_ "fourth-exam/api_gateway_evrone/api/docs"
	"github.com/casbin/casbin/v2"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type RouteOption struct {
	Config         *config.Config
	Logger         *zap.Logger
	ContextTimeout time.Duration
	Cache          redis.Cache
	Enforcer       *casbin.CachedEnforcer
	Service        grpserviceclient.ServiceClient
	BrokerProducer event.BrokerProducer
}

// NewRoute
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func NewRoute(option RouteOption) http.Handler {
	handleOption := &handlers.HandlerOption{
		Config:         option.Config,
		Logger:         option.Logger,
		ContextTimeout: option.ContextTimeout,
		Cache:          option.Cache,
		Enforcer:       option.Enforcer,
		Service:        option.Service,
		BrokerProducer: option.BrokerProducer,
	}

	router := chi.NewRouter()
	router.Use(chimiddleware.RealIP, chimiddleware.Logger, chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(option.ContextTimeout))
	router.Use(cors.Handler(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-Id"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Route("/v1", func(r chi.Router) {
		r.Use(middleware.AuthContext(option.Config.Token.Secret))
		r.Mount("/user", v1.NewUserHandler(handleOption))
	})

	// declare swagger api route
	router.Get("/swagger/*", httpSwagger.Handler())
	return router
}
