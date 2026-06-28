package httpapi_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func bootstrapAndResolve(t *testing.T, store *httpapi.PostgresStore, server http.Handler, token, platformServerID, platformUserID string) (playerID, communityID string) {
	t.Helper()

	// Bootstrap identity
	bootRec := doJSON(server, http.MethodGet, "/v1/me", token, nil)
	if bootRec.Code != http.StatusOK {
		t.Fatalf("bootstrap identity: expected 200, got %d: %s", bootRec.Code, bootRec.Body.String())
	}

	// Resolve external actor
	result, err := store.ResolveExternalActor(context.Background(), "discord", platformServerID, platformUserID, "Test Player", "Test Community")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}
	return result.PlayerID, result.CommunityID
}

func createRound(t *testing.T, server http.Handler, store *httpapi.PostgresStore, token, communityID string) (roundID string) {
	t.Helper()

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

	startBody := map[string]string{
		"communityId":       communityID,
		"revealTimeframeId": tfs.Timeframes[0].ID,
	}
	startRec := doJSON(server, http.MethodPost, "/v1/rounds", token, startBody)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("start round: expected 201, got %d: %s", startRec.Code, startRec.Body.String())
	}

	type startResp struct {
		Round struct {
			ID string `json:"id"`
		} `json:"round"`
	}
	var sr startResp
	decodeResponse(t, startRec, &sr)
	return sr.Round.ID
}

func TestCommitToRoundSuccess(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithStore(store)
	_, communityID := bootstrapAndResolve(t, store, server, "alice-token", "srv-commit-1", "alice-dc")
	roundID := createRound(t, server, store, "alice-token", communityID)

	rec := doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "alice-token", nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		CommitID string `json:"commitId"`
	}
	decodeResponse(t, rec, &resp)
	if resp.CommitID == "" {
		t.Fatal("expected non-empty commitId")
	}
}

func TestCommitToRoundRequiresAuth(t *testing.T) {
	server := newTestServer(t)
	rec := doJSON(server, http.MethodPost, "/v1/rounds/some-round/commits", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCommitToRoundFailsForNonexistentRound(t *testing.T) {
	server := newTestServer(t)
	rec := doJSON(server, http.MethodPost, "/v1/rounds/nonexistent-round/commits", "alice-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCommitToRoundFailsWhenAlreadyCommitted(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithStore(store)
	_, communityID := bootstrapAndResolve(t, store, server, "alice-token", "srv-commit-2", "alice-dc-2")
	roundID := createRound(t, server, store, "alice-token", communityID)

	// First commit succeeds
	firstRec := doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "alice-token", nil)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("first commit: expected 201, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	// Second commit fails
	secondRec := doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "alice-token", nil)
	if secondRec.Code != http.StatusForbidden {
		t.Fatalf("second commit: expected 403, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
}

func TestSubmitJumpSuccess(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithStore(store)
	_, communityID := bootstrapAndResolve(t, store, server, "alice-token", "srv-submit-1", "alice-dc-s1")
	roundID := createRound(t, server, store, "alice-token", communityID)

	// Commit first
	commitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "alice-token", nil)
	if commitRec.Code != http.StatusCreated {
		t.Fatalf("commit: expected 201, got %d: %s", commitRec.Code, commitRec.Body.String())
	}

	// Submit
	submitBody := map[string]any{
		"caption":      "Crunchwrap at the opera",
		"evidenceUrls": []string{"https://example.com/photo1.jpg"},
	}
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/jumps", "alice-token", submitBody)
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit: expected 201, got %d: %s", submitRec.Code, submitRec.Body.String())
	}

	var resp struct {
		Jump struct {
			ID           string   `json:"id"`
			Caption      string   `json:"caption"`
			EvidenceURLs []string `json:"evidenceUrls"`
		} `json:"jump"`
	}
	decodeResponse(t, submitRec, &resp)
	if resp.Jump.ID == "" {
		t.Fatal("expected non-empty jump id")
	}
	if resp.Jump.Caption != "Crunchwrap at the opera" {
		t.Fatalf("expected caption preserved, got %q", resp.Jump.Caption)
	}
	if len(resp.Jump.EvidenceURLs) != 1 {
		t.Fatalf("expected 1 evidence URL, got %d", len(resp.Jump.EvidenceURLs))
	}
}

func TestSubmitJumpFailsWithoutCommit(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithStore(store)
	_, communityID := bootstrapAndResolve(t, store, server, "alice-token", "srv-submit-2", "alice-dc-s2")
	roundID := createRound(t, server, store, "alice-token", communityID)

	submitBody := map[string]string{
		"caption": "Should fail — not committed",
	}
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/jumps", "alice-token", submitBody)
	if submitRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", submitRec.Code, submitRec.Body.String())
	}
}

func TestSubmitJumpFailsWithoutCaption(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithStore(store)
	_, communityID := bootstrapAndResolve(t, store, server, "alice-token", "srv-submit-3", "alice-dc-s3")
	roundID := createRound(t, server, store, "alice-token", communityID)

	// Commit
	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "alice-token", nil)

	// Submit without caption
	submitBody := map[string]any{
		"evidenceUrls": []string{"https://example.com/photo.jpg"},
	}
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/jumps", "alice-token", submitBody)
	if submitRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", submitRec.Code, submitRec.Body.String())
	}
}

func TestListJumpsSealedPreReveal(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithStore(store)
	_, communityID := bootstrapAndResolve(t, store, server, "alice-token", "srv-seal-1", "alice-dc-seal1")
	bootstrapAndResolve(t, store, server, "bob-token", "srv-seal-1", "bob-dc-seal1")
	roundID := createRound(t, server, store, "alice-token", communityID)

	// Alice commits and submits
	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "alice-token", nil)
	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/jumps", "alice-token", map[string]any{
		"caption":      "Alice's secret jump",
		"evidenceUrls": []string{"https://example.com/alice-secret.jpg"},
	})

	// Bob commits and submits too
	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "bob-token", nil)
	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/jumps", "bob-token", map[string]any{
		"caption":      "Bob's secret jump",
		"evidenceUrls": []string{"https://example.com/bob-secret.jpg"},
	})

	// Alice views jumps — should see her own unsealed, Bob's sealed
	rec := doJSON(server, http.MethodGet, "/v1/rounds/"+roundID+"/jumps", "alice-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Jumps []struct {
			Caption      string `json:"caption"`
			SealedViewer bool   `json:"sealedViewer"`
		} `json:"jumps"`
		CommitCount    int `json:"commitCount"`
		SubmissionCount int `json:"submissionCount"`
	}
	decodeResponse(t, rec, &resp)

	if len(resp.Jumps) != 2 {
		t.Fatalf("expected 2 jumps, got %d", len(resp.Jumps))
	}
	if resp.CommitCount != 2 {
		t.Fatalf("expected commitCount=2, got %d", resp.CommitCount)
	}
	if resp.SubmissionCount != 2 {
		t.Fatalf("expected submissionCount=2, got %d", resp.SubmissionCount)
	}

	// Exactly one jump should be unsealed (Alice's), one sealed (Bob's)
	sealedCount := 0
	unsealedCount := 0
	for _, j := range resp.Jumps {
		if j.SealedViewer {
			sealedCount++
			if j.Caption != "" {
				t.Fatalf("sealed jump should have empty caption, got %q", j.Caption)
			}
		} else {
			unsealedCount++
			if j.Caption == "" {
				t.Fatal("unsealed jump should have caption")
			}
		}
	}
	if sealedCount != 1 {
		t.Fatalf("expected 1 sealed jump, got %d", sealedCount)
	}
	if unsealedCount != 1 {
		t.Fatalf("expected 1 unsealed jump, got %d", unsealedCount)
	}
}

func TestListJumpsShowsGhostJumper(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithStore(store)
	_, communityID := bootstrapAndResolve(t, store, server, "alice-token", "srv-ghost-1", "alice-dc-ghost1")
	bootstrapAndResolve(t, store, server, "bob-token", "srv-ghost-1", "bob-dc-ghost1")
	roundID := createRound(t, server, store, "alice-token", communityID)

	// Alice commits but does NOT submit (ghost jumper)
	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "alice-token", nil)

	// Bob commits and submits
	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "bob-token", nil)
	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/jumps", "bob-token", map[string]any{
		"caption": "Bob's jump",
	})

	// Alice views jumps — should see herself as committed-but-not-submitted
	rec := doJSON(server, http.MethodGet, "/v1/rounds/"+roundID+"/jumps", "alice-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Jumps          []struct {
			PlayerHasCommitted  bool `json:"playerHasCommitted"`
			PlayerHasSubmitted  bool `json:"playerHasSubmitted"`
		} `json:"jumps"`
		CommitCount    int `json:"commitCount"`
		SubmissionCount int `json:"submissionCount"`
	}
	decodeResponse(t, rec, &resp)

	if resp.CommitCount != 2 {
		t.Fatalf("expected commitCount=2, got %d", resp.CommitCount)
	}
	if resp.SubmissionCount != 1 {
		t.Fatalf("expected submissionCount=1, got %d", resp.SubmissionCount)
	}
}

func TestGetJumpAuthorSeesFull(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithStore(store)
	_, communityID := bootstrapAndResolve(t, store, server, "alice-token", "srv-getj-1", "alice-dc-get1")
	roundID := createRound(t, server, store, "alice-token", communityID)

	doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/commits", "alice-token", nil)
	rec := doJSON(server, http.MethodPost, "/v1/rounds/"+roundID+"/jumps", "alice-token", map[string]any{
		"caption":      "My jump",
		"evidenceUrls": []string{"https://example.com/pic.jpg"},
	})

	var submitResp struct {
		Jump struct {
			ID string `json:"id"`
		} `json:"jump"`
	}
	decodeResponse(t, rec, &submitResp)

	getRec := doJSON(server, http.MethodGet, "/v1/rounds/"+roundID+"/jumps/"+submitResp.Jump.ID, "alice-token", nil)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	var getResp struct {
		Jump struct {
			Caption      string `json:"caption"`
			SealedViewer bool   `json:"sealedViewer"`
		} `json:"jump"`
	}
	decodeResponse(t, getRec, &getResp)

	if getResp.Jump.SealedViewer {
		t.Fatal("author should see unsealed content")
	}
	if getResp.Jump.Caption != "My jump" {
		t.Fatalf("expected caption, got %q", getResp.Jump.Caption)
	}
}

func TestListJumpsRequiresAuth(t *testing.T) {
	server := newTestServer(t)
	rec := doJSON(server, http.MethodGet, "/v1/rounds/some-round/jumps", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
