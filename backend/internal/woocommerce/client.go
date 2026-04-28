// Package woocommerce wraps the WooCommerce REST API v3.
package woocommerce

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/univerbeauty777/univer-tracker/backend/pkg/httpclient"
)

// Client talks to a single WooCommerce store.
type Client struct {
	baseURL        string
	consumerKey    string
	consumerSecret string
	http           *http.Client
}

// New creates a WooCommerce client. baseURL example: https://lizzon.com.br
func New(baseURL, consumerKey, consumerSecret string) *Client {
	return &Client{
		baseURL:        strings.TrimRight(baseURL, "/"),
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		http:           httpclient.Default(),
	}
}

// do builds an authenticated request against /wp-json/wc/v3/{path} and decodes the response.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	endpoint := fmt.Sprintf("%s/wp-json/wc/v3/%s", c.baseURL, strings.TrimLeft(path, "/"))

	if query == nil {
		query = url.Values{}
	}
	query.Set("consumer_key", c.consumerKey)
	query.Set("consumer_secret", c.consumerSecret)
	endpoint += "?" + query.Encode()

	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = strings.NewReader(string(raw))
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpclient.DoWithRetry(ctx, c.http, req, 3)
	if err != nil {
		return fmt.Errorf("woocommerce %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("woocommerce %s %s: http %d: %s", method, path, resp.StatusCode, string(raw))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
