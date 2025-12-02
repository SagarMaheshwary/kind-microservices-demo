package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/constant"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/helper"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/logger"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/service"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/transports/http/client"
)

type UserHandlerOpts struct {
	UserService service.UserService
	Logger      logger.Logger
	HTTPClient  *http.Client
}

type userHandler struct {
	userService service.UserService
	logger      logger.Logger
	httpClient  *http.Client
}

type CreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateUserValidationError struct {
	Name     []string `json:"name"`
	Email    []string `json:"email"`
	Password []string `json:"password"`
}

func NewUserHandler(opts *UserHandlerOpts) *userHandler {
	return &userHandler{
		userService: opts.UserService,
		logger:      opts.Logger,
		httpClient:  opts.HTTPClient,
	}
}

func (u *userHandler) CreateUser(c *gin.Context) {
	var in CreateUserInput
	if err := c.ShouldBindJSON(&in); err != nil {
		res := helper.PrepareResponseFromValidationError(err, &CreateUserValidationError{})
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res, err := u.userService.CreateUser(c.Request.Context(), &client.CreateUserRequest{
		Name:     in.Name,
		Email:    in.Email,
		Password: in.Password,
	})

	if err != nil && errors.Is(err, constant.ErrHTTPServiceUnavailable) {
		c.JSON(http.StatusServiceUnavailable, helper.PrepareResponse(constant.ErrHTTPServiceUnavailable.Error(), nil))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.PrepareResponse("error", nil))
		return
	}

	c.JSON(http.StatusOK, helper.PrepareResponse("ok", gin.H{"user": res}))
}
