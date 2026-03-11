package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/heyashy/bb/internal/auth"
)

const defaultBaseURL = "https://api.bitbucket.org/2.0"

type Client interface {
	Get(ctx context.Context, path string, query url.Values) (*http.Response, error)
	Post(ctx context.Context, path string, body interface{}) (*http.Response, error)
	Put(ctx context.Context, path string, body interface{}) (*http.Response, error)
	Delete(ctx context.Context, path string) (*http.Response, error)
}

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	auth       auth.Provider
}

func NewClient(authProvider auth.Provider) *HTTPClient {
	return &HTTPClient{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		auth: authProvider,
	}
}

func (c *HTTPClient) Get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	u := c.baseURL + path
	if query != nil {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("infra: cannot create request: %w", err)
	}

	return c.do(req)
}

func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.doWithBody(ctx, http.MethodPost, path, body)
}

func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.doWithBody(ctx, http.MethodPut, path, body)
}

func (c *HTTPClient) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("infra: cannot create request: %w", err)
	}

	return c.do(req)
}

func (c *HTTPClient) doWithBody(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("infra: cannot marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("infra: cannot create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.do(req)
}

func (c *HTTPClient) do(req *http.Request) (*http.Response, error) {
	if err := c.auth.ApplyAuth(req); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("infra: request failed: %w", err)
	}

	return resp, nil
}

func DecodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("domain: API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	if v == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func ReadRawBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("domain: API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("infra: cannot read response body: %w", err)
	}

	return string(body), nil
}
