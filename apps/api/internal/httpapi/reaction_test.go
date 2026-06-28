package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func TestListStampCatalogReturnsSeeded(t *testing.T) {
	server := newTestServer(t)

	rec := doJSON(server, http.MethodGet, "/v1/stamp-catalog", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	type stampDTO struct {
		ID     string `json:"id"`
		Stance string `json:"stance"`
		Label  string `json:"label"`
		Glyph  string `json:"glyph"`
	}
	type catalogResp struct {
		Stamps []stampDTO `json:"stamps"`
	}
	var resp catalogResp
	decodeResponse(t, rec, &resp)

	if len(resp.Stamps) != 6 {
		t.Fatalf("expected 6 stamps, got %d", len(resp.Stamps))
	}

	stances := map[string]bool{}
	for _, s := range resp.Stamps {
		if s.ID == "" {
			t.Fatal("stamp ID must not be empty")
		}
		if s.Stance == "" {
			t.Fatal("stamp stance must not be empty")
		}
		if s.Label == "" {
			t.Fatal("stamp label must not be empty")
		}
		stances[s.Stance] = true
	}
	if len(stances) != 6 {
		t.Fatalf("expected 6 unique stances, got %d", len(stances))
	}
}

func TestApplyReactionAppliesStampToRevealedJump(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	// Freeze clock for round creation, then advance for reveal
	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
			"bob-token":   {Provider: "test-provider", Subject: "bob-auth", Email: "bob@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap alice
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	server.ServeHTTP(bootRec, req)
	if bootRec.Code != http.StatusOK {
		t.Fatalf("bootstrap alice: %d", bootRec.Code)
	}

	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-react", "alice-discord-react", "Alice", "Test Community React")
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

	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": tfs.Timeframes[0].ID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: expected 201, got %d", startRec.Code)
	}

	type roundDTO struct {
		ID string `json:"id"`
	}
	type startResp struct {
		Round roundDTO `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	// Commit and submit as alice
	commitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "alice-token", nil)
	if commitRec.Code != http.StatusCreated {
		t.Fatalf("commit: %d", commitRec.Code)
	}

	jumpBody := map[string]any{
		"caption":      "Test jump",
		"evidenceUrls": []string{"https://example.com/pic.jpg"},
	}
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps", "alice-token", jumpBody)
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit: expected 201, got %d: %s", submitRec.Code, submitRec.Body.String())
	}

	// Get the jump ID from the submit response
	type jumpDTO struct {
		ID string `json:"id"`
	}
	type submitResp struct {
		Jump jumpDTO `json:"jump"`
	}
	var subResp submitResp
	decodeResponse(t, submitRec, &subResp)

	// Advance clock past reveal_by and reveal
	store.SetClock(func() time.Time { return past.Add(25 * time.Hour) })
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("reveal: expected 200, got %d: %s", revealRec.Code, revealRec.Body.String())
	}

	// Get stamp ID for "approval" stance
	stampRec := doJSON(server, http.MethodGet, "/v1/stamp-catalog", "", nil)
	type stampDTO struct {
		ID     string `json:"id"`
		Stance string `json:"stance"`
	}
	type catalogResp struct {
		Stamps []stampDTO `json:"stamps"`
	}
	var cat catalogResp
	decodeResponse(t, stampRec, &cat)

	var approvalStampID string
	for _, s := range cat.Stamps {
		if s.Stance == "approval" {
			approvalStampID = s.ID
			break
		}
	}
	if approvalStampID == "" {
		t.Fatal("could not find approval stamp")
	}

	// Apply reaction
	reactBody := map[string]string{"stampId": approvalStampID}
	reactRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/reactions", "bob-token", reactBody)
	if reactRec.Code != http.StatusCreated {
		t.Fatalf("apply reaction: expected 201, got %d: %s", reactRec.Code, reactRec.Body.String())
	}

	type reactionDTO struct {
		ID       string `json:"id"`
		StampID  string `json:"stampId"`
		JumpID   string `json:"jumpId"`
		PlayerID string `json:"playerId"`
	}
	type reactResp struct {
		Reaction reactionDTO `json:"reaction"`
	}
	var rr reactResp
	decodeResponse(t, reactRec, &rr)

	if rr.Reaction.StampID != approvalStampID {
		t.Fatalf("expected StampID %s, got %s", approvalStampID, rr.Reaction.StampID)
	}
	if rr.Reaction.JumpID != subResp.Jump.ID {
		t.Fatalf("expected JumpID %s, got %s", subResp.Jump.ID, rr.Reaction.JumpID)
	}
}

func TestApplyReactionFailsOnNonRevealedRound(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time { return time.Now().UTC() })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	server.ServeHTTP(bootRec, req)

	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-react-2", "alice-discord-react-2", "Alice", "Test Community React 2")
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

	// Find long timeframe to ensure round not revealed
	var longTFID string
	for _, tf := range tfs.Timeframes {
		if tf.ID != "" {
			longTFID = tf.ID
			break
		}
	}

	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": longTFID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: expected 201, got %d", startRec.Code)
	}

	type roundDTO struct {
		ID string `json:"id"`
	}
	type startResp struct {
		Round roundDTO `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	// Commit and submit
	commitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "alice-token", nil)
	if commitRec.Code != http.StatusCreated {
		t.Fatalf("commit: %d", commitRec.Code)
	}

	jumpBody := map[string]any{
		"caption":      "Test jump",
		"evidenceUrls": []string{"https://example.com/pic.jpg"},
	}
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps", "alice-token", jumpBody)
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit: %d", submitRec.Code)
	}

	type jumpDTO struct {
		ID string `json:"id"`
	}
	type submitResp struct {
		Jump jumpDTO `json:"jump"`
	}
	var subResp submitResp
	decodeResponse(t, submitRec, &subResp)

	// Get stamp
	stampRec := doJSON(server, http.MethodGet, "/v1/stamp-catalog", "", nil)
	type stampDTO struct {
		ID     string `json:"id"`
		Stance string `json:"stance"`
	}
	type catalogResp struct {
		Stamps []stampDTO `json:"stamps"`
	}
	var cat catalogResp
	decodeResponse(t, stampRec, &cat)

	var approvalStampID string
	for _, s := range cat.Stamps {
		if s.Stance == "approval" {
			approvalStampID = s.ID
			break
		}
	}

	// Try reaction on non-revealed round
	reactBody := map[string]string{"stampId": approvalStampID}
	reactRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/reactions", "alice-token", reactBody)
	if reactRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on non-revealed round, got %d: %s", reactRec.Code, reactRec.Body.String())
	}
}

func TestApplyReactionFailsOnUnknownStamp(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	server.ServeHTTP(bootRec, req)

	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-react-3", "alice-discord-react-3", "Alice", "Test Community React 3")
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

	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": tfs.Timeframes[0].ID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: %d", startRec.Code)
	}

	type roundDTO struct {
		ID string `json:"id"`
	}
	type startResp struct {
		Round roundDTO `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	// Commit and submit
	commitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "alice-token", nil)
	if commitRec.Code != http.StatusCreated {
		t.Fatalf("commit: %d", commitRec.Code)
	}

	jumpBody := map[string]any{
		"caption":      "Test jump",
		"evidenceUrls": []string{"https://example.com/pic.jpg"},
	}
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps", "alice-token", jumpBody)
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit: %d", submitRec.Code)
	}

	type jumpDTO struct {
		ID string `json:"id"`
	}
	type submitResp struct {
		Jump jumpDTO `json:"jump"`
	}
	var subResp submitResp
	decodeResponse(t, submitRec, &subResp)

	// Reveal
	store.SetClock(func() time.Time { return past.Add(25 * time.Hour) })
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("reveal: expected 200, got %d: %s", revealRec.Code, revealRec.Body.String())
	}

	// Try reaction with unknown stamp
	reactBody := map[string]string{"stampId": "stamp_nonexistent"}
	reactRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/reactions", "alice-token", reactBody)
	if reactRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on unknown stamp, got %d: %s", reactRec.Code, reactRec.Body.String())
	}
}

func TestApplyReactionFailsOnUnknownJump(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	server.ServeHTTP(bootRec, req)

	// Get stamp
	stampRec := doJSON(server, http.MethodGet, "/v1/stamp-catalog", "", nil)
	type stampDTO struct {
		ID string `json:"id"`
	}
	type catalogResp struct {
		Stamps []stampDTO `json:"stamps"`
	}
	var cat catalogResp
	decodeResponse(t, stampRec, &cat)

	reactBody := map[string]string{"stampId": cat.Stamps[0].ID}
	reactRec := doJSON(server, http.MethodPost, "/v1/rounds/round-nonexistent/jumps/jump-nonexistent/reactions", "alice-token", reactBody)
	if reactRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on unknown jump, got %d: %s", reactRec.Code, reactRec.Body.String())
	}
}

func TestApplyReactionRequiresAuth(t *testing.T) {
	server := newTestServer(t)

	reactBody := map[string]string{"stampId": "stamp-1"}
	reactRec := doJSON(server, http.MethodPost, "/v1/rounds/round-1/jumps/jump-1/reactions", "", reactBody)
	if reactRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", reactRec.Code)
	}
}

func TestApplyReactionFailsOnMissingStampID(t *testing.T) {
	server := newTestServer(t)

	reactBody := map[string]string{}
	reactRec := doJSON(server, http.MethodPost, "/v1/rounds/round-1/jumps/jump-1/reactions", "alice-token", reactBody)
	if reactRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on missing stampId, got %d: %s", reactRec.Code, reactRec.Body.String())
	}
}

func TestApplyReactionAllowsMultipleStamps(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"bob-token": {Provider: "test-provider", Subject: "bob-auth", Email: "bob@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap bob
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer bob-token")
	server.ServeHTTP(bootRec, req)

	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-react-multi", "bob-discord-multi", "Bob", "Test Community Multi")
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

	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": tfs.Timeframes[0].ID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "bob-token", startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: %d", startRec.Code)
	}

	type roundDTO struct {
		ID string `json:"id"`
	}
	type startResp struct {
		Round roundDTO `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	// Commit and submit
	commitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "bob-token", nil)
	if commitRec.Code != http.StatusCreated {
		t.Fatalf("commit: %d", commitRec.Code)
	}

	jumpBody := map[string]any{
		"caption":      "Multi-stamp jump",
		"evidenceUrls": []string{"https://example.com/pic.jpg"},
	}
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps", "bob-token", jumpBody)
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit: %d", submitRec.Code)
	}

	type jumpDTO struct {
		ID string `json:"id"`
	}
	type submitResp struct {
		Jump jumpDTO `json:"jump"`
	}
	var subResp submitResp
	decodeResponse(t, submitRec, &subResp)

	// Reveal
	store.SetClock(func() time.Time { return past.Add(25 * time.Hour) })
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "bob-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("reveal: %d", revealRec.Code)
	}

	// Get all stamps
	stampRec := doJSON(server, http.MethodGet, "/v1/stamp-catalog", "", nil)
	type stampDTO struct {
		ID string `json:"id"`
	}
	type catalogResp struct {
		Stamps []stampDTO `json:"stamps"`
	}
	var cat catalogResp
	decodeResponse(t, stampRec, &cat)

	// Apply two different stamps
	for i := 0; i < 2 && i < len(cat.Stamps); i++ {
		reactBody := map[string]string{"stampId": cat.Stamps[i].ID}
		reactRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/reactions", "bob-token", reactBody)
		if reactRec.Code != http.StatusCreated {
			t.Fatalf("reaction %d: expected 201, got %d: %s", i, reactRec.Code, reactRec.Body.String())
		}
	}

	// Same stamp again — should be forbidden (duplicate)
	reactBody := map[string]string{"stampId": cat.Stamps[0].ID}
	dupRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/reactions", "bob-token", reactBody)
	if dupRec.Code != http.StatusForbidden {
		t.Fatalf("duplicate reaction: expected 403, got %d: %s", dupRec.Code, dupRec.Body.String())
	}
}
