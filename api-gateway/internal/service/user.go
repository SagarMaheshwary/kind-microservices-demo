package service

import (
	"context"
	"fmt"

	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/transports/http/client"
)

type UserService interface {
	CreateUser(ctx context.Context, user *client.CreateUserRequest) (*client.CreateUserResponse, error)
}

type userService struct {
	httpClient *client.UserClient
}

func NewUserService(httpClient *client.UserClient) *userService {
	return &userService{
		httpClient: httpClient,
	}
}

func (u *userService) CreateUser(ctx context.Context, user *client.CreateUserRequest) (*client.CreateUserResponse, error) {
	res, err := u.httpClient.CreateUser(ctx, &client.CreateUserRequest{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return res, nil
}
