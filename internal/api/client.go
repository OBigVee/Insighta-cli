package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"insighta-cli/internal/auth"
)

// Client wraps HTTP calls with auth token injection and auto-refresh
type Client struct {
	httpClient *http.Client
	backendURL string
	creds      *auth.Credentials
}

func NewClient() *Client {
	creds, err := auth.LoadCredentials()
	if err != nil {
		fmt.Println("❌ Not logged in. Run 'insighta login' first.")
		return &Client{
			httpClient: &http.Client{Timeout: 30 * time.Second},
		}
	}

	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		backendURL: creds.BackendURL,
		creds:      creds,
	}
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.creds == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	req.Header.Set("Authorization", "Bearer "+c.creds.AccessToken)
	req.Header.Set("X-API-Version", "1")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If 401, try refreshing token and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		if err := auth.RefreshTokens(); err != nil {
			return nil, fmt.Errorf("session expired. Run 'insighta login'")
		}

		// Reload credentials after refresh
		newCreds, err := auth.LoadCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to reload credentials")
		}
		c.creds = newCreds

		// Retry with new token
		req.Header.Set("Authorization", "Bearer "+c.creds.AccessToken)
		return c.httpClient.Do(req)
	}

	return resp, nil
}

func (c *Client) Get(path string) (*http.Response, error) {
	url := c.backendURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

func (c *Client) Post(path string, body string) (*http.Response, error) {
	url := c.backendURL + path
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}
