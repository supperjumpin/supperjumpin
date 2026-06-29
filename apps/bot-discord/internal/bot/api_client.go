package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Actor struct {
	GuildID string
	UserID  string
	Tuple   string
}

type ActorResolver interface {
	Resolve(ctx context.Context, guildID, userID string) (Actor, error)
}

type ActorResolverFunc func(ctx context.Context, guildID, userID string) (Actor, error)

func (f ActorResolverFunc) Resolve(ctx context.Context, guildID, userID string) (Actor, error) {
	return f(ctx, guildID, userID)
}

type APIClientConfig struct {
	BaseURL      string
	AdapterToken string
	HTTPClient   *http.Client
}

type APIClient struct {
	baseURL      string
	adapterToken string
	httpClient   *http.Client
}

func NewAPIClient(cfg APIClientConfig) (*APIClient, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("APIClientConfig.BaseURL is required")
	}
	if cfg.AdapterToken == "" {
		return nil, fmt.Errorf("APIClientConfig.AdapterToken is required")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &APIClient{
		baseURL:      cfg.BaseURL,
		adapterToken: cfg.AdapterToken,
		httpClient:   httpClient,
	}, nil
}

func (c *APIClient) CommitToRound(ctx context.Context, roundID string, actor Actor) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/rounds/%s/commits", c.baseURL, roundID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build commit request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.adapterToken)
	httpReq.Header.Set("X-Adapter-Actor", actor.Tuple)
	return c.httpClient.Do(httpReq)
}

func (c *APIClient) EvaluateReveal(ctx context.Context, roundID string, actor Actor) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/rounds/%s/reveal", c.baseURL, roundID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build reveal request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.adapterToken)
	httpReq.Header.Set("X-Adapter-Actor", actor.Tuple)
	return c.httpClient.Do(httpReq)
}

func (c *APIClient) GetRecap(ctx context.Context, roundID string, actor Actor) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/rounds/%s/recap", c.baseURL, roundID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build recap request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.adapterToken)
	httpReq.Header.Set("X-Adapter-Actor", actor.Tuple)
	return c.httpClient.Do(httpReq)
}

func (c *APIClient) ListRoundJumps(ctx context.Context, roundID string, actor Actor) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/rounds/%s/jumps", c.baseURL, roundID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build list jumps request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.adapterToken)
	httpReq.Header.Set("X-Adapter-Actor", actor.Tuple)
	return c.httpClient.Do(httpReq)
}

func (c *APIClient) ListStampCatalog(ctx context.Context) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/stamp-catalog", nil)
	if err != nil {
		return nil, fmt.Errorf("build stamp catalog request: %w", err)
	}
	return c.httpClient.Do(httpReq)
}

func (c *APIClient) PostRoundComment(ctx context.Context, roundID, body string, actor Actor) (*http.Response, error) {
	return c.postComment(ctx, fmt.Sprintf("%s/v1/rounds/%s/comments", c.baseURL, roundID), body, actor)
}

func (c *APIClient) PostJumpComment(ctx context.Context, roundID, jumpID, body string, actor Actor) (*http.Response, error) {
	return c.postComment(ctx, fmt.Sprintf("%s/v1/rounds/%s/jumps/%s/comments", c.baseURL, roundID, jumpID), body, actor)
}

func (c *APIClient) postComment(ctx context.Context, url, body string, actor Actor) (*http.Response, error) {
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return nil, fmt.Errorf("marshal comment: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build comment request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.adapterToken)
	httpReq.Header.Set("X-Adapter-Actor", actor.Tuple)
	httpReq.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(httpReq)
}

func (c *APIClient) StartRound(ctx context.Context, req StartRoundRequestBody, actor Actor) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal start round request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/rounds", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build start round request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.adapterToken)
	httpReq.Header.Set("X-Adapter-Actor", actor.Tuple)
	httpReq.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(httpReq)
}

func (c *APIClient) ApplyReaction(ctx context.Context, roundID, jumpID, stampID string, actor Actor) (*http.Response, error) {
	body, err := json.Marshal(map[string]string{"stampId": stampID})
	if err != nil {
		return nil, fmt.Errorf("marshal apply reaction: %w", err)
	}
	url := fmt.Sprintf("%s/v1/rounds/%s/jumps/%s/reactions", c.baseURL, roundID, jumpID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build apply reaction request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.adapterToken)
	httpReq.Header.Set("X-Adapter-Actor", actor.Tuple)
	httpReq.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(httpReq)
}

func (c *APIClient) SubmitJump(ctx context.Context, roundID, caption string, evidenceURLs []string, actor Actor) (*http.Response, error) {
	body, err := json.Marshal(map[string]any{
		"caption":      caption,
		"evidenceUrls": evidenceURLs,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal submit jump: %w", err)
	}
	url := fmt.Sprintf("%s/v1/rounds/%s/jumps", c.baseURL, roundID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build submit jump request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.adapterToken)
	httpReq.Header.Set("X-Adapter-Actor", actor.Tuple)
	httpReq.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(httpReq)
}

type StartRoundRequestBody struct {
	CommunityID       string `json:"communityId"`
	PromptID          string `json:"promptId,omitempty"`
	RevealTimeframeID string `json:"revealTimeframeId"`
}
