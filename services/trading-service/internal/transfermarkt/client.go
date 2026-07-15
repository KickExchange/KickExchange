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
	ExternalID    string
	Name          string
	MarketValue   int64
	ImageURL      string
	Position      string
	Club          string
	Nationalities []string
	ShirtNumber   string
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
		ImageURL    string `json:"imageUrl"`
		ShirtNumber string `json:"shirtNumber"`
		Position    struct {
			Main string `json:"main"`
		} `json:"position"`
		Club struct {
			Name string `json:"name"`
		} `json:"club"`
		Citizenship []string `json:"citizenship"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return PlayerProfile{}, fmt.Errorf("transfermarkt: decode response: %w", err)
	}

	return PlayerProfile{
		ExternalID:    raw.ID,
		Name:          raw.Name,
		MarketValue:   raw.MarketValue,
		ImageURL:      raw.ImageURL,
		Position:      raw.Position.Main,
		Club:          raw.Club.Name,
		Nationalities: raw.Citizenship,
		ShirtNumber:   raw.ShirtNumber,
	}, nil
}

// FullProfile is everything the profile endpoint returns - richer than
// PlayerProfile, used for the player detail page rather than search/preview.
type FullProfile struct {
	ExternalID          string
	Name                string
	Description         string
	NameInHomeCountry   string
	ImageURL            string
	PlaceOfBirthCity    string
	PlaceOfBirthCountry string
	Height              int
	Citizenship         []string
	Position            string
	PositionOther       []string
	Foot                string
	ShirtNumber         string
	ClubName            string
	ClubJoined          string
	ClubContractExpires string
	MarketValue         int64
	AgentName           string
	Outfitter           string
	SocialMedia         []string
}

func (c *Client) GetFullProfile(ctx context.Context, externalID string) (FullProfile, error) {
	reqURL := fmt.Sprintf("%s/players/%s/profile", c.baseURL, externalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return FullProfile{}, fmt.Errorf("transfermarkt: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return FullProfile{}, fmt.Errorf("transfermarkt: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return FullProfile{}, ErrPlayerNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return FullProfile{}, fmt.Errorf("transfermarkt: unexpected status %d", resp.StatusCode)
	}

	var raw struct {
		ID                string `json:"id"`
		Name              string `json:"name"`
		Description       string `json:"description"`
		NameInHomeCountry string `json:"nameInHomeCountry"`
		ImageURL          string `json:"imageUrl"`
		PlaceOfBirth      struct {
			City    string `json:"city"`
			Country string `json:"country"`
		} `json:"placeOfBirth"`
		Height      int      `json:"height"`
		Citizenship []string `json:"citizenship"`
		Position    struct {
			Main  string   `json:"main"`
			Other []string `json:"other"`
		} `json:"position"`
		Foot        string `json:"foot"`
		ShirtNumber string `json:"shirtNumber"`
		Club        struct {
			Name            string `json:"name"`
			Joined          string `json:"joined"`
			ContractExpires string `json:"contractExpires"`
		} `json:"club"`
		MarketValue int64 `json:"marketValue"`
		Agent       struct {
			Name string `json:"name"`
		} `json:"agent"`
		Outfitter   string   `json:"outfitter"`
		SocialMedia []string `json:"socialMedia"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return FullProfile{}, fmt.Errorf("transfermarkt: decode response: %w", err)
	}

	return FullProfile{
		ExternalID:          raw.ID,
		Name:                raw.Name,
		Description:         raw.Description,
		NameInHomeCountry:   raw.NameInHomeCountry,
		ImageURL:            raw.ImageURL,
		PlaceOfBirthCity:    raw.PlaceOfBirth.City,
		PlaceOfBirthCountry: raw.PlaceOfBirth.Country,
		Height:              raw.Height,
		Citizenship:         raw.Citizenship,
		Position:            raw.Position.Main,
		PositionOther:       raw.Position.Other,
		Foot:                raw.Foot,
		ShirtNumber:         raw.ShirtNumber,
		ClubName:            raw.Club.Name,
		ClubJoined:          raw.Club.Joined,
		ClubContractExpires: raw.Club.ContractExpires,
		MarketValue:         raw.MarketValue,
		AgentName:           raw.Agent.Name,
		Outfitter:           raw.Outfitter,
		SocialMedia:         raw.SocialMedia,
	}, nil
}

type Transfer struct {
	Date         string
	Season       string
	ClubFromName string
	ClubToName   string
	MarketValue  int64
}

func (c *Client) GetTransfers(ctx context.Context, externalID string) ([]Transfer, error) {
	reqURL := fmt.Sprintf("%s/players/%s/transfers", c.baseURL, externalID)
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
		Transfers []struct {
			Date     string `json:"date"`
			Season   string `json:"season"`
			ClubFrom struct {
				Name string `json:"name"`
			} `json:"clubFrom"`
			ClubTo struct {
				Name string `json:"name"`
			} `json:"clubTo"`
			MarketValue int64 `json:"marketValue"`
		} `json:"transfers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("transfermarkt: decode response: %w", err)
	}

	transfers := make([]Transfer, len(raw.Transfers))
	for i, t := range raw.Transfers {
		transfers[i] = Transfer{
			Date:         t.Date,
			Season:       t.Season,
			ClubFromName: t.ClubFrom.Name,
			ClubToName:   t.ClubTo.Name,
			MarketValue:  t.MarketValue,
		}
	}
	return transfers, nil
}

// SearchPlayers looks players up by name, most relevant match first per
// transfermarkt's own ranking. The search endpoint returns less than the
// profile endpoint - no image or shirt number.
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
			Position    string `json:"position"`
			Club        struct {
				Name string `json:"name"`
			} `json:"club"`
			Nationalities []string `json:"nationalities"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("transfermarkt: decode response: %w", err)
	}

	profiles := make([]PlayerProfile, len(raw.Results))
	for i, r := range raw.Results {
		profiles[i] = PlayerProfile{
			ExternalID:    r.ID,
			Name:          r.Name,
			MarketValue:   r.MarketValue,
			Position:      r.Position,
			Club:          r.Club.Name,
			Nationalities: r.Nationalities,
		}
	}
	return profiles, nil
}
