package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func TestEvaluateRevealFiresWhenTimePassed(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	// Freeze clock for round creation
	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap identity
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	server.ServeHTTP(bootRec, req)
	if bootRec.Code != http.StatusOK {
		t.Fatalf("bootstrap identity: expected 200, got %d", bootRec.Code)
	}

	// Create community
	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-reveal-1", "alice-discord", "Alice", "Test Community Reveal")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}

	// Get timeframe (24h)
	tfRec := doJSON(server, http.MethodGet, "/v1/reveal-timeframes", "", nil)
	type tfResp struct {
		Timeframes []struct {
			ID string `json:"id"`
		} `json:"timeframes"`
	}
	var tfs tfResp
	decodeResponse(t, tfRec, &tfs)

	// Start a round with 24h timeframe; reveal_by = past + 24h
	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": tfs.Timeframes[0].ID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "alice-token", startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: expected 201, got %d: %s", startRec.Code, startRec.Body.String())
	}

	type roundDTO struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	type startResp struct {
		Round roundDTO `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	// Advance clock past reveal_by for the reveal evaluation
	store.SetClock(func() time.Time { return past.Add(25 * time.Hour) })

	// Evaluate reveal — now should fire
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("evaluate reveal: expected 200, got %d: %s", revealRec.Code, revealRec.Body.String())
	}

	type revealResp struct {
		Round    roundDTO `json:"round"`
		Revealed bool     `json:"revealed"`
	}
	var rr revealResp
	decodeResponse(t, revealRec, &rr)

	if !rr.Revealed {
		t.Fatal("expected Revealed=true when time has passed")
	}
	if rr.Round.Status != "revealed" {
		t.Fatalf("expected Round.Status=revealed, got %s", rr.Round.Status)
	}
}

func TestEvaluateRevealReturnsNotRevealedWhenBeforeTime(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"bob-token": {Provider: "test-provider", Subject: "bob-auth", Email: "bob@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap identity
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer bob-token")
	server.ServeHTTP(bootRec, req)
	if bootRec.Code != http.StatusOK {
		t.Fatalf("bootstrap identity: expected 200, got %d", bootRec.Code)
	}

	// Create community
	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-reveal-2", "bob-discord", "Bob", "Test Community Reveal 2")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}

	// Get timeframe (7 days = 168h)
	tfRec := doJSON(server, http.MethodGet, "/v1/reveal-timeframes", "", nil)
	type tfResp struct {
		Timeframes []struct {
			ID            string `json:"id"`
			DurationHours int    `json:"durationHours"`
		} `json:"timeframes"`
	}
	var tfs tfResp
	decodeResponse(t, tfRec, &tfs)

	// Find the 168h timeframe
	var longTFID string
	for _, tf := range tfs.Timeframes {
		if tf.DurationHours == 168 {
			longTFID = tf.ID
			break
		}
	}
	if longTFID == "" {
		t.Fatal("could not find 168h timeframe")
	}

	// Start a round with the 7-day timeframe (reveal is far in future relative to real clock)
	startBody := map[string]string{
		"communityId":       result.CommunityID,
		"revealTimeframeId": longTFID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", "bob-token", startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: expected 201, got %d: %s", startRec.Code, startRec.Body.String())
	}

	type roundDTO struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	type startResp struct {
		Round roundDTO `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	// Evaluate reveal — should NOT be revealed yet (clock is real, 7 days not passed)
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "bob-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("evaluate reveal: expected 200, got %d: %s", revealRec.Code, revealRec.Body.String())
	}

	type revealResp struct {
		Round    roundDTO `json:"round"`
		Revealed bool     `json:"revealed"`
	}
	var rr revealResp
	decodeResponse(t, revealRec, &rr)

	if rr.Revealed {
		t.Fatal("expected Revealed=false when before reveal time")
	}
	if rr.Round.Status != "active" {
		t.Fatalf("expected Round.Status=active, got %s", rr.Round.Status)
	}
}

func TestEvaluateRevealIdempotent(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	// Freeze clock for round creation
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

	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-reveal-3", "alice-discord-3", "Alice", "Test Community Reveal 3")
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
		t.Fatalf("start round: expected 201, got %d: %s", startRec.Code, startRec.Body.String())
	}

	type roundDTO struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	type startResp struct {
		Round roundDTO `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)

	// Advance clock past reveal_by
	store.SetClock(func() time.Time { return past.Add(25 * time.Hour) })

	// First reveal
	rec1 := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first reveal: expected 200, got %d: %s", rec1.Code, rec1.Body.String())
	}

	type revealResp struct {
		Round    roundDTO `json:"round"`
		Revealed bool     `json:"revealed"`
	}
	var rr1 revealResp
	decodeResponse(t, rec1, &rr1)
	if !rr1.Revealed {
		t.Fatal("first reveal: expected Revealed=true")
	}
	if rr1.Round.Status != "revealed" {
		t.Fatalf("first reveal: expected Status=revealed, got %s", rr1.Round.Status)
	}

	// Second reveal — idempotent
	rec2 := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if rec2.Code != http.StatusOK {
		t.Fatalf("second reveal: expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var rr2 revealResp
	decodeResponse(t, rec2, &rr2)
	if !rr2.Revealed {
		t.Fatal("second reveal: expected Revealed=true (idempotent)")
	}
	if rr2.Round.Status != "revealed" {
		t.Fatalf("second reveal: expected Status=revealed, got %s", rr2.Round.Status)
	}
}

func TestEvaluateRevealRoundNotFound(t *testing.T) {
	server := newTestServer(t)

	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/nonexistent/reveal", "alice-token", nil)
	if revealRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for nonexistent round, got %d: %s", revealRec.Code, revealRec.Body.String())
	}
}

func TestEvaluateRevealRequiresAuth(t *testing.T) {
	server := newTestServer(t)

	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/round-1/reveal", "", nil)
	if revealRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated request, got %d", revealRec.Code)
	}
}

func TestJumpContentVisibleAfterReveal(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	// Freeze clock for round creation
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

	// Resolve alice to create community + player
	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-reveal-seal", "alice-discord-seal", "Alice", "Test Community Seal")
	if err != nil {
		t.Fatalf("resolve alice: %v", err)
	}

	// Resolve bob to create his player record
	_, err = store.ResolveExternalActor(context.Background(), "discord", "server-reveal-seal", "bob-discord-seal", "Bob", "Test Community Seal")
	if err != nil {
		t.Fatalf("resolve bob: %v", err)
	}

	tfRec := doJSON(server, http.MethodGet, "/v1/reveal-timeframes", "", nil)
	type tfResp struct {
		Timeframes []struct {
			ID string `json:"id"`
		} `json:"timeframes"`
	}
	var tfs tfResp
	decodeResponse(t, tfRec, &tfs)

	// Start a round
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

	// Bootstrap bob
	bootBobRec := httptest.NewRecorder()
	reqBob := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	reqBob.Header.Set("Authorization", "Bearer bob-token")
	server.ServeHTTP(bootBobRec, reqBob)
	if bootBobRec.Code != http.StatusOK {
		t.Fatalf("bootstrap bob: %d", bootBobRec.Code)
	}

	// Commit and submit as alice
	commitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "alice-token", nil)
	if commitRec.Code != http.StatusCreated {
		t.Fatalf("commit: expected 201, got %d: %s", commitRec.Code, commitRec.Body.String())
	}

	jumpBody := map[string]any{
		"caption":      "My secret jump",
		"evidenceUrls": []string{"https://example.com/secret.jpg"},
	}
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps", "alice-token", jumpBody)
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit: expected 201, got %d: %s", submitRec.Code, submitRec.Body.String())
	}

	// Before reveal: bob looks at list — should see sealed content (clock still at past, reveal not yet due)
	listRec := doJSON(server, http.MethodGet, "/v1/rounds/"+sr.Round.ID+"/jumps", "bob-token", nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list jumps pre-reveal: expected 200, got %d", listRec.Code)
	}

	type jumpDTO struct {
		Caption      string `json:"caption"`
		SealedViewer bool   `json:"sealedViewer"`
	}
	type listResp struct {
		Jumps []jumpDTO `json:"jumps"`
	}
	var lr listResp
	decodeResponse(t, listRec, &lr)
	if len(lr.Jumps) == 0 {
		t.Fatal("expected at least one jump in list")
	}
	if !lr.Jumps[0].SealedViewer {
		t.Fatal("pre-reveal: bob should see sealed jump")
	}
	if lr.Jumps[0].Caption != "" {
		t.Fatalf("pre-reveal: bob should not see caption, got %q", lr.Jumps[0].Caption)
	}

	// Advance clock past reveal_by and reveal the round
	store.SetClock(func() time.Time { return past.Add(25 * time.Hour) })
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("reveal: expected 200, got %d: %s", revealRec.Code, revealRec.Body.String())
	}

	// After reveal: bob looks at list — should see unsealed content
	listRec2 := doJSON(server, http.MethodGet, "/v1/rounds/"+sr.Round.ID+"/jumps", "bob-token", nil)
	if listRec2.Code != http.StatusOK {
		t.Fatalf("list jumps post-reveal: expected 200, got %d", listRec2.Code)
	}

	var lr2 listResp
	decodeResponse(t, listRec2, &lr2)
	if len(lr2.Jumps) == 0 {
		t.Fatal("expected at least one jump in list after reveal")
	}
	if lr2.Jumps[0].SealedViewer {
		t.Fatal("post-reveal: bob should see unsealed jump")
	}
	if lr2.Jumps[0].Caption != "My secret jump" {
		t.Fatalf("post-reveal: bob should see caption, got %q", lr2.Jumps[0].Caption)
	}
}
