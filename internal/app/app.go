package app

import (
	"context"
	"fmt"
	"fourth-exam/api_gateway_evrone/api"
	grpserviceclient "fourth-exam/api_gateway_evrone/internal/infrastructure/grp_service_client"
	"fourth-exam/api_gateway_evrone/internal/infrastructure/kafka"
	redisrepo "fourth-exam/api_gateway_evrone/internal/infrastructure/repository/redis"
	"fourth-exam/api_gateway_evrone/internal/pkg/config"
	"fourth-exam/api_gateway_evrone/internal/pkg/logger"
	"fourth-exam/api_gateway_evrone/internal/pkg/policy"
	"fourth-exam/api_gateway_evrone/internal/pkg/postgres"
	"fourth-exam/api_gateway_evrone/internal/pkg/redis"
	"fourth-exam/api_gateway_evrone/internal/usecase/event"
	"net/http"
	"time"

	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
)

type App struct {
	Config         *config.Config
	Logger         *zap.Logger
	DB             *postgres.PostgresDB
	RedisDB        *redis.RedisDB
	server         *http.Server
	Enforcer       *casbin.CachedEnforcer
	Clients        grpserviceclient.ServiceClient
	BrokerProducer event.BrokerProducer
}

func NewApp(cfg config.Config) (*App, error) {
	// logger init
	logger, err := logger.New(cfg.LogLevel, cfg.Environment, cfg.APP+".log")
	if err != nil {
		return nil, err
	}

	// kafka producer init
	kafkaProducer := kafka.NewProducer(&cfg, logger)

	// postgres init
	db, err := postgres.New(&cfg)
	if err != nil {
		return nil, err
	}

	// redis init
	redisdb, err := redis.New(&cfg)
	if err != nil {
		return nil, err
	}

	// initialization enforcer
	enforcer, err := policy.NewCachedEnforcer(&cfg, logger)
	if err != nil {
		return nil, err
	}

	enforcer.SetCache(policy.NewCache(&redisdb.Client))

	return &App{
		Config:         &cfg,
		Logger:         logger,
		DB:             db,
		RedisDB:        redisdb,
		Enforcer:       enforcer,
		BrokerProducer: kafkaProducer,
	}, nil
}

func (a *App) Run() error {
	contextTimeout, err := time.ParseDuration(a.Config.Context.Timeout)
	if err != nil {
		return fmt.Errorf("error while parsing context timeout: %v", err)
	}

	clients, err := grpserviceclient.New(a.Config)
	if err != nil {
		return err
	}
	a.Clients = clients

	// initialize cache
	cache := redisrepo.NewCache(a.RedisDB)

	// api init
	handler := api.NewRoute(api.RouteOption{
		Config:         a.Config,
		Logger:         a.Logger,
		ContextTimeout: contextTimeout,
		Cache:          cache,
		Enforcer:       a.Enforcer,
		Service:        clients,
		BrokerProducer: a.BrokerProducer,
	})
	if err = a.Enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("error during enforcer load policy: %w", err)
	}

	// server init
	a.server, err = api.NewServer(a.Config, handler)
	if err != nil {
		return fmt.Errorf("error while initializing server: %v", err)
	}

	return a.server.ListenAndServe()
}

func (a *App) Stop() {

	// close database
	a.DB.Close()

	// close grpc connections
	a.Clients.Close()

	// kafka producer close
	a.BrokerProducer.Close()

	// shutdown server http
	if err := a.server.Shutdown(context.Background()); err != nil {
		a.Logger.Error("shutdown server http ", zap.Error(err))
	}

	// zap logger sync
	a.Logger.Sync()
}
