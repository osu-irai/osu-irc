package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ClientConfig struct {
	secret string
	name   string
}

type ApiClient struct {
	config ClientConfig
}

func CreateClientConfig() ApiClient {
	return ApiClient{
		config: ClientConfig{
			secret: os.Getenv("IRC_CLIENT_SECRET"),
			name:   os.Getenv("IRC_CLIENT_NAME"),
		},
	}
}

func (c *ApiClient) GetUserIdList() ([]int64, error) {
	resp, err := http.Get("IRC_API_ENDPOINT")

	if err != nil {
		return nil, fmt.Errorf("failed to make an API request, %v", err)
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to parse response, %v", err)
	}

	var uids []int64
	err = json.Unmarshal(data, &uids)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response, %v", err)
	}
	return uids, nil
}
