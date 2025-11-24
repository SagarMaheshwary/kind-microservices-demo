package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type UserClient struct {
	http *HTTPClient
}

func NewUserClient(httpClient *HTTPClient) *UserClient {
	return &UserClient{http: httpClient}
}

func (c *UserClient) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.http.BaseURL+"/users",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New("http api error")
	}

	var response CreateUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
