package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const clientTimeout = 5 * time.Second

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: clientTimeout,
		},
	}
}

func (c *Client) DoJSON(
	method, path string,
	query url.Values,
	reqBody any,
	wantStatus int,
	respBody any,
) (int, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var body []byte
	var err error

	if reqBody != nil {
		body, err = json.Marshal(reqBody)
		if err != nil {
			return 0, fmt.Errorf("marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, u, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if wantStatus != 0 && resp.StatusCode != wantStatus {
		return resp.StatusCode, fmt.Errorf("unexpected status: got %d, want %d", resp.StatusCode, wantStatus)
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return resp.StatusCode, fmt.Errorf("decode response: %w", err)
		}
	}

	return resp.StatusCode, nil
}
