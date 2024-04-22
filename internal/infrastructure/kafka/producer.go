package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"fourth-exam/api_gateway_evrone/internal/pkg/config"

	pb "fourth-exam/api_gateway_evrone/genproto/user_service"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type producer struct {
	logger      *zap.Logger
	userService *kafka.Writer
}

func NewProducer(config *config.Config, logger *zap.Logger) *producer {
	return &producer{
		logger: logger,
		userService: &kafka.Writer{
			Addr:                   kafka.TCP(config.Kafka.Address...),
			Topic:                  config.Kafka.Topic.UserService,
			Balancer:               &kafka.Hash{},
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
			Async:                  true,
			Completion: func(messages []kafka.Message, err error) {
				if err != nil {
					logger.Error("kafka userCreated", zap.Error(err))
				}
				for _, message := range messages {
					logger.Sugar().Info(
						"kafka userCreated message",
						zap.Int("partition", message.Partition),
						zap.Int64("offset", message.Offset),
						zap.String("key", string(message.Key)),
						zap.String("value", string(message.Value)),
					)
				}
			},
		},
	}
}

func (p *producer) buildMessage(key string, value []byte) kafka.Message {
	return kafka.Message{
		Key:   []byte(key),
		Value: value,
	}
}

func (p *producer) ProduceUserInfoToKafka(ctx context.Context, key string, body *pb.User) error {
	// data, err := body.Marshal()
	// if err != nil {
	// 	return err
	// }
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	fmt.Println(body)
	message := p.buildMessage(key, data)

	return p.userService.WriteMessages(ctx, message)
}

func (p *producer) Close() {
	if err := p.userService.Close(); err != nil {
		p.logger.Error("error during close writer userCreated", zap.Error(err))
	}
}
