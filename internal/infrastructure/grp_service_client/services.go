package grpserviceclient

import (
	"fmt"
	pb "fourth-exam/api_gateway_evrone/genproto/user_service"
	"fourth-exam/api_gateway_evrone/internal/pkg/config"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

type ServiceClient interface {
	UserService() pb.UserServiceClient
	Close()
}

type serviceClient struct {
	connections []*grpc.ClientConn
	userService pb.UserServiceClient
}

func New(cfg *config.Config) (ServiceClient, error) {
	connUserService, err := grpc.Dial(
		fmt.Sprintf("%s%s", cfg.UserService.Host, cfg.UserService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &serviceClient{
		userService: pb.NewUserServiceClient(connUserService),
		connections: []*grpc.ClientConn{
			connUserService,
		},
	}, nil
}

func (s *serviceClient) UserService() pb.UserServiceClient {
	return s.userService
}

func (s *serviceClient) Close() {
	for _, conn := range s.connections {
		if err := conn.Close(); err != nil {
			log.Printf("error while closing grpc connection: %v", err)
		}
	}
}