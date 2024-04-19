package event

import (
	"context"
	pb "fourth-exam/api_gateway_evrone/genproto/user_service"
)

type ConsumerConfig interface {
	GetBrokers() []string
	GetTopic() string
	GetGroupID() string
	GetHandler() ConsumerHandler
}

type ConsumerHandler interface {
	Handle(ctx context.Context, key, value []byte) error
}

type BrokerConsumer interface {
	Run() error
	RegisterConsumer(cfg ConsumerConfig)
	Close()
}

type BrokerProducer interface {
	ProduceUserInfoToKafka(ctx context.Context, key string, body *pb.User) error
	Close()
}
