package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func TestRevealTimeframesReturnsSeeded(t *testing.T) {
	server := newTestServer(t)
	rec := doJSON(server, http.MethodGet, "/v1/reveal-timeframes", "", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	type tfDTO struct {
		ID            string `json:"id"`
		Label         string `json:"label"`
		DurationHours int    `json:"durationHours"`
	}
	type response struct {
		Timeframes []tfDTO `json:"timeframes"`
	}
	var resp response
	decodeResponse(t, rec, &resp)

	if len(resp.Timeframes) != 3 {
		t.Fatalf("expected 3 seeded timeframes, got %d", len(resp.Timeframes))
	}
	if resp.Timeframes[0].DurationHours != 24 {
		t.Fatalf("expected first timeframe 24 hours, got %d", resp.Timeframes[0].DurationHours)
	}
	if resp.Timeframes[1].DurationHours != 72 {
		t.Fatalf("expected second timeframe 72 hours, got %d", resp.Timeframes[1].DurationHours)
	}
	if resp.Timeframes[2].DurationHours != 168 {
		t.Fatalf("expected third timeframe 168 hours, got %d", resp.Timeframes[2].DurationHours)
	}
}

func TestStartRoundCreatesRoundWithRandomPrompt(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap identity to create player
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("X-Adapter-Actor", "discord:test-server:alice-user")
	server.ServeHTTP(bootRec, req)
	if bootRec.Code != http.StatusOK {
		t.Fatalf("bootstrap identity: expected 200, got %d: %s", bootRec.Code, bootRec.Body.String())
	}

	// Create community via ResolveExternalActor
	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}

	// Get timeframe
	tfRec := doJSON(server, http.MethodGet, "/v1/reveal-timeframes", "", nil)
	type tfResp struct {
		Timeframes []struct {
			ID string `json:"id"`
		} `json:"timeframes"`
	}
	var tfs tfResp
	decodeResponse(t, tfRec, &tfs)
	if len(tfs.Timeframes) == 0 {
		t.Fatal("expected at least one timeframe")
	}

	// Start round without promptId (random)
	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": tfs.Timeframes[0].ID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: expected 201, got %d: %s", startRec.Code, startRec.Body.String())
	}

	type roundDTO struct {
		ID          string `json:"id"`
		CommunityID string `json:"communityId"`
		PromptID    string `json:"promptId"`
		Status      string `json:"status"`
		Prompt      *struct {
			ID   string `json:"id"`
			Copy string `json:"copy"`
		} `json:"prompt"`
	}
	type startResp struct {
		Round roundDTO `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	if sr.Round.ID == "" {
		t.Fatal("expected round id")
	}
	if sr.Round.CommunityID != result.CommunityID {
		t.Fatalf("expected community %s, got %s", result.CommunityID, sr.Round.CommunityID)
	}
	if sr.Round.Status != "active" {
		t.Fatalf("expected status active, got %s", sr.Round.Status)
	}
	if sr.Round.PromptID == "" {
		t.Fatal("expected prompt id from random selection")
	}
	if sr.Round.Prompt == nil {
		t.Fatal("expected prompt details in start round response")
	}
	if sr.Round.Prompt.ID != sr.Round.PromptID {
		t.Fatalf("expected prompt.id %s, got %s", sr.Round.PromptID, sr.Round.Prompt.ID)
	}
	if sr.Round.Prompt.Copy == "" {
		t.Fatal("expected prompt copy in start round response")
	}
}

func TestStartRoundFailsWhenActiveRoundExists(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap identity
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("X-Adapter-Actor", "discord:test-server:alice-user")
	server.ServeHTTP(bootRec, req)

	// Create community
	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community 2")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}

	// Get timeframe
	tfRec := doJSON(server, http.MethodGet, "/v1/reveal-timeframes", "", nil)
	type tfResp struct {
		Timeframes []struct {
			ID string `json:"id"`
		} `json:"timeframes"`
	}
	var tfs tfResp
	decodeResponse(t, tfRec, &tfs)

	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": tfs.Timeframes[0].ID,
	}

	// First round succeeds
	firstRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("first start round: expected 201, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	// Second round fails with 403
	secondRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if secondRec.Code != http.StatusForbidden {
		t.Fatalf("second start round: expected 403, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
}

func TestStartRoundRequiresAuth(t *testing.T) {
	server := newTestServer(t)
	startBody := map[string]string{
		"communityId":       "irrelevant",
		"revealTimeframeId": "irrelevant",
	}
	rec := doJSON(server, http.MethodPost, "/v1/rounds", "", startBody)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestStartRoundRejectsMissingCommunityId(t *testing.T) {
	server := newTestServer(t)
	startBody := map[string]string{
		"revealTimeframeId": "irrelevant",
	}
	rec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStartRoundRejectsMissingTimeframeId(t *testing.T) {
	server := newTestServer(t)
	startBody := map[string]string{
		"communityId": "irrelevant",
	}
	rec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStartRoundRejectsInvalidTimeframe(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {},
		},
		Store: store,
		Now:   store.Now,
	})

	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("X-Adapter-Actor", "discord:test-server:alice-user")
	server.ServeHTTP(bootRec, req)

	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community 3")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}

	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": "nonexistent",
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if startRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", startRec.Code, startRec.Body.String())
	}
}

func TestStartRoundWithExplicitPrompt(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {},
		},
		Store: store,
		Now:   store.Now,
	})

	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("X-Adapter-Actor", "discord:test-server:alice-user")
	server.ServeHTTP(bootRec, req)

	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community 4")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}

	tfRec := doJSON(server, http.MethodGet, "/v1/reveal-timeframes", "", nil)
	type tfResp struct {
		Timeframes []struct {
			ID string `json:"id"`
		} `json:"timeframes"`
	}
	var tfs tfResp
	decodeResponse(t, tfRec, &tfs)

	// Get prompt catalog to get a valid prompt ID
	catRec := doJSON(server, http.MethodGet, "/v1/prompt-catalog", "", nil)
	var cat struct {
		Packs []struct {
			Prompts []struct {
				ID string `json:"id"`
			} `json:"prompts"`
		} `json:"packs"`
	}
	decodeResponse(t, catRec, &cat)
	if len(cat.Packs) == 0 || len(cat.Packs[0].Prompts) == 0 {
		t.Fatal("expected at least one prompt in catalog")
	}
	promptID := cat.Packs[0].Prompts[0].ID

	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": tfs.Timeframes[0].ID,
		"promptId":          promptID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: expected 201, got %d: %s", startRec.Code, startRec.Body.String())
	}

	type startResp struct {
		Round struct {
			PromptID string `json:"promptId"`
			Prompt   *struct {
				ID   string `json:"id"`
				Copy string `json:"copy"`
			} `json:"prompt"`
		} `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	if sr.Round.PromptID != promptID {
		t.Fatalf("expected PromptID %s, got %s", promptID, sr.Round.PromptID)
	}
	if sr.Round.Prompt == nil {
		t.Fatal("expected prompt details in explicit start round response")
	}
	if sr.Round.Prompt.ID != promptID {
		t.Fatalf("expected prompt.id %s, got %s", promptID, sr.Round.Prompt.ID)
	}
	if sr.Round.Prompt.Copy == "" {
		t.Fatal("expected prompt copy in explicit start round response")
	}
}
