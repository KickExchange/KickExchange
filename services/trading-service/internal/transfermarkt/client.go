// Package transfermarkt calls the self-hosted transfermarkt-api container
// for player identity data (github.com/felipeall/transfermarkt-api).
package transfermarkt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var ErrPlayerNotFound = fmt.Errorf("transfermarkt: player not found")

type PlayerProfile struct {
	ExternalID  string
	Name        string
	MarketValue int64
}

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) GetPlayerProfile(ctx context.Context, externalID string) (PlayerProfile, error) {
	reqURL := fmt.Sprintf("%s/players/%s/profile", c.baseURL, externalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return PlayerProfile{}, fmt.Errorf("transfermarkt: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return PlayerProfile{}, fmt.Errorf("transfermarkt: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return PlayerProfile{}, ErrPlayerNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return PlayerProfile{}, fmt.Errorf("transfermarkt: unexpected status %d", resp.StatusCode)
	}

	var raw struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		MarketValue int64  `json:"marketValue"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return PlayerProfile{}, fmt.Errorf("transfermarkt: decode response: %w", err)
	}

	return PlayerProfile{ExternalID: raw.ID, Name: raw.Name, MarketValue: raw.MarketValue}, nil
}

// SearchPlayers looks players up by name, most relevant match first per
// transfermarkt's own ranking.
func (c *Client) SearchPlayers(ctx context.Context, name string) ([]PlayerProfile, error) {
	reqURL := fmt.Sprintf("%s/players/search/%s?page_number=1", c.baseURL, url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("transfermarkt: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transfermarkt: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transfermarkt: unexpected status %d", resp.StatusCode)
	}

	var raw struct {
		Results []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			MarketValue int64  `json:"marketValue"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("transfermarkt: decode response: %w", err)
	}

	profiles := make([]PlayerProfile, len(raw.Results))
	for i, r := range raw.Results {
		profiles[i] = PlayerProfile{ExternalID: r.ID, Name: r.Name, MarketValue: r.MarketValue}
	}
	return profiles, nil
}
