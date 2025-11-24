package service

import (
	"context"
	"time"

	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/constant"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/logger"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/rabbitmq"
)

type UserService interface {
	Create(ctx context.Context, user *User) (*User, error)
}

type userService struct {
	rabbitmq rabbitmq.RabbitMQ
	logger   logger.Logger
	users    []*User
}

type UserServiceOpts struct {
	Rabbitmq rabbitmq.RabbitMQ
	Logger   logger.Logger
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

func NewUserService(opts *UserServiceOpts) *userService {
	return &userService{
		rabbitmq: opts.Rabbitmq,
		logger:   opts.Logger,
		users:    []*User{},
	}
}

func (u *userService) Create(ctx context.Context, user *User) (*User, error) {
	//Pretend DB query
	time.Sleep(100 * time.Millisecond)

	user.ID = 1
	if len(u.users) > 0 {
		user.ID = u.users[len(u.users)-1].ID + 1
	}
	user.Password = ""

	u.users = append(u.users, user)

	err := u.rabbitmq.Publish(ctx, constant.QUEUE_NOTIFICATION_SERVICE, &rabbitmq.MessageType{
		Pattern: constant.EVENT_USER_CREATED,
		Data:    user,
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}
