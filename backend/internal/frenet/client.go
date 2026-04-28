// Package frenet wraps the Frenet tracking API.
package frenet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/univerbeauty777/univer-tracker/backend/pkg/httpclient"
)

const baseURL = "http://api.frenet.com.br"

// Client talks to the Frenet REST API.
type Client struct {
	token string
	http  *http.Client
}

// New returns a new Frenet client authenticated with the given token.
func New(token string) *Client {
	return &Client{
		token: token,
		http:  httpclient.Default(),
	}
}

// TrackingEvent is a single carrier event in a tracking timeline.
type TrackingEvent struct {
	EventDateTime    string `json:"EventDateTime"`
	EventLocation    string `json:"EventLocation"`
	EventDescription string `json:"EventDescription"`
	EventType        string `json:"EventType"`
}

// TrackingResponse is the Frenet trackinginfo response shape.
type TrackingResponse struct {
	TrackingNumber     string          `json:"TrackingNumber"`
	TrackingURL        string          `json:"TrackingUrl"`
	ServiceDescription string          `json:"ServiceDescrition"` // Frenet typo
	ErrorMessage       string          `json:"ErrorMessage"`
	TrackingEvents     []TrackingEvent `json:"TrackingEvents"`
}

// GetTrackingInfo queries Frenet for events of a given tracking number.
// shippingServiceCode is required (e.g. "03298" for PAC, "03220" for Sedex).
func (c *Client) GetTrackingInfo(ctx context.Context, trackingNumber, shippingServiceCode string) (*TrackingResponse, error) {
	clean := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(trackingNumber), " ", ""))

	body := map[string]string{
		"TrackingNumber":      clean,
		"ShippingServiceCode": shippingServiceCode,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/tracking/trackinginfo", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", c.token)

	resp, err := httpclient.DoWithRetry(ctx, c.http, req, 3)
	if err != nil {
		return nil, fmt.Errorf("frenet trackinginfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("frenet trackinginfo: http %d: %s", resp.StatusCode, string(raw))
	}

	var out TrackingResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
