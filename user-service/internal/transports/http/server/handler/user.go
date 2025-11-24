package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/logger"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/service"
)

type UserHandlerOpts struct {
	UserService service.UserService
	Logger      logger.Logger
}

type userHandler struct {
	userService service.UserService
	logger      logger.Logger
}

type CreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewUserHandler(opts *UserHandlerOpts) *userHandler {
	return &userHandler{
		userService: opts.UserService,
		logger:      opts.Logger,
	}
}

func (u *userHandler) CreateUser(c *gin.Context) {
	var in CreateUserInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	user, _ := u.userService.Create(c.Request.Context(), &service.User{
		Name:     in.Name,
		Email:    in.Email,
		Password: in.Password,
	})

	c.JSON(http.StatusOK, user)
}
