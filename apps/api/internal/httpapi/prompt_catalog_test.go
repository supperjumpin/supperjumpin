package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPromptCatalogReturnsSeededCatalogWithNestedPrompts(t *testing.T) {
	server := newTestServer(t)

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/prompt-catalog", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Packs []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
			Description string `json:"description"`
			Prompts     []struct {
				ID       string `json:"id"`
				Copy     string `json:"copy"`
				Theme    string `json:"theme"`
				CostTier string `json:"costTier"`
			} `json:"prompts"`
		} `json:"packs"`
	}
	decodeResponse(t, rec, &body)

	if len(body.Packs) != 3 {
		t.Fatalf("expected 3 seeded packs, got %d", len(body.Packs))
	}

	packByID := make(map[string]struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
		Prompts     []struct {
			ID       string `json:"id"`
			Copy     string `json:"copy"`
			Theme    string `json:"theme"`
			CostTier string `json:"costTier"`
		} `json:"prompts"`
	})
	for _, p := range body.Packs {
		packByID[p.ID] = p
	}

	kitchen, ok := packByID["prompt_pack_8c22a060a7c2"]
	if !ok {
		t.Fatalf("expected Kitchen Classics pack to be present, got packs %v", packByID)
	}
	if kitchen.DisplayName != "Kitchen Classics" {
		t.Fatalf("expected displayName 'Kitchen Classics', got %q", kitchen.DisplayName)
	}
	if kitchen.Description == "" {
		t.Fatalf("expected non-empty description for Kitchen Classics")
	}
	if len(kitchen.Prompts) != 4 {
		t.Fatalf("expected 4 prompts in Kitchen Classics, got %d", len(kitchen.Prompts))
	}

	for _, p := range kitchen.Prompts {
		if p.CostTier != "tier_1" {
			t.Fatalf("expected costTier 'tier_1' (snake_case preserved over the wire), got %q on prompt %s", p.CostTier, p.ID)
		}
		if p.Theme == "" {
			t.Fatalf("expected non-empty theme on prompt %s", p.ID)
		}
	}
}

func TestGetPromptCatalogAllPromptsAreReachable(t *testing.T) {
	server := newTestServer(t)

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/prompt-catalog", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Packs []struct {
			Prompts []struct {
				ID string `json:"id"`
			} `json:"prompts"`
		} `json:"packs"`
	}
	decodeResponse(t, rec, &body)

	seen := make(map[string]bool)
	for _, p := range body.Packs {
		for _, prompt := range p.Prompts {
			seen[prompt.ID] = true
		}
	}
	if len(seen) != 9 {
		t.Fatalf("expected 9 seeded prompts across packs, got %d unique prompts", len(seen))
	}
}

func TestGetPromptCatalogRejectsMethodOtherThanGet(t *testing.T) {
	server := newTestServer(t)

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/prompt-catalog", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d: %s", rec.Code, rec.Body.String())
	}
}
