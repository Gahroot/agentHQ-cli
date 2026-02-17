package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/agenthq/cli/internal/common/config"
)

type APIResponse struct {
	Success    bool            `json:"success"`
	Data       json.RawMessage `json:"data,omitempty"`
	Error      *APIError       `json:"error,omitempty"`
	Pagination *Pagination     `json:"pagination,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Pagination struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasMore bool `json:"hasMore"`
}

type Client struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
}

func New() (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &Client{
		baseURL:    cfg.HubURL,
		authToken:  cfg.GetAuthToken(),
		httpClient: &http.Client{},
	}, nil
}

func NewWithToken(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		authToken:  token,
		httpClient: &http.Client{},
	}
}

func (c *Client) Request(method, path string, body interface{}, query map[string]string) (*APIResponse, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}

	if query != nil {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return &apiResp, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return &apiResp, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return &apiResp, nil
}

func (c *Client) Get(path string, query map[string]string) (*APIResponse, error) {
	return c.Request("GET", path, nil, query)
}

func (c *Client) Post(path string, body interface{}) (*APIResponse, error) {
	return c.Request("POST", path, body, nil)
}

func (c *Client) Patch(path string, body interface{}) (*APIResponse, error) {
	return c.Request("PATCH", path, body, nil)
}

func (c *Client) Delete(path string) (*APIResponse, error) {
	return c.Request("DELETE", path, nil, nil)
}
