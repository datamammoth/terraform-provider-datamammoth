package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	APIKey  string
	BaseURL string
	HTTP    *http.Client
}

func New(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://app.datamammoth.com/api/v2"
	}
	return &Client{APIKey: apiKey, BaseURL: baseURL, HTTP: &http.Client{Timeout: 60 * time.Second}}
}

func (c *Client) do(method, path string, body interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %v", resp.StatusCode, result)
	}
	return result, nil
}

func (c *Client) Get(path string) (map[string]interface{}, error) {
	return c.do("GET", path, nil)
}

func (c *Client) Post(path string, body interface{}) (map[string]interface{}, error) {
	return c.do("POST", path, body)
}

func (c *Client) Patch(path string, body interface{}) (map[string]interface{}, error) {
	return c.do("PATCH", path, body)
}

func (c *Client) Delete(path string) (map[string]interface{}, error) {
	return c.do("DELETE", path, nil)
}
