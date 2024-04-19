package handlers

import (
	"context"
	"fourth-exam/api_gateway_evrone/api/middleware"
	grpserviceclient "fourth-exam/api_gateway_evrone/internal/infrastructure/grp_service_client"
	"fourth-exam/api_gateway_evrone/internal/infrastructure/repository/redis"
	"fourth-exam/api_gateway_evrone/internal/pkg/config"
	"fourth-exam/api_gateway_evrone/internal/usecase/event"
	"time"

	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
)

const (
	UserToken = "user"
)

type HandlerOption struct {
	Config *config.Config
	Logger *zap.Logger
	ContextTimeout time.Duration
	Enforcer *casbin.CachedEnforcer
	Cache redis.Cache
	Service grpserviceclient.ServiceClient
	BrokerProducer event.BrokerProducer
}

type BaseHandler struct {
	Cache redis.Cache
	Config *config.Config
	Client grpserviceclient.ServiceClient
}

func (h *BaseHandler) GetAuthData(ctx context.Context) (map[string]string, bool) {
	data, ok := ctx.Value(middleware.RequestAuthCtx).(map[string]string)
	return data, ok
}