package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type ClientConfig struct {
	Secret string `json:"ClientSecret"`
	Name   string `json:"Name"`
}

type ApiClient struct {
	config ClientConfig
	token  string
}

func CreateClientConfig() ApiClient {
	return ApiClient{
		config: ClientConfig{
			Secret: os.Getenv("IRC_CLIENT_SECRET"),
			Name:   "Irc",
		},
		token: "",
	}
}

func (c *ApiClient) PrepareAuthToken() error {
	body, err := json.Marshal(c.config)
	if err != nil {
		return fmt.Errorf("failed to prepare request body, %v", err)
	}

	req, err := http.NewRequest("POST", os.Getenv("IRC_API_AUTH"), bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("failed to prepare request, %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make an API request, %v", err)
	}
	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to parse response, %v", err)
	}

	c.token = string(data)

	return nil

}

func (c *ApiClient) GetUserIdList() ([]string, error) {
	err := c.PrepareAuthToken()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare string, %v", err)
	}
	var uids []string

	req, err := http.NewRequest("GET", os.Getenv("IRC_API_USERNAMES"), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", c.token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make an API request, %v", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed, %v", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to parse response, %v", err)
	}
	err = json.Unmarshal(data, &uids)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal usernames, %v, %v", err, data)
	}
	log.Printf("found %d usernames", len(uids))
	return uids, nil
}
