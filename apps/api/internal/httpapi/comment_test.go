package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func TestPostCommentOnRevealedJump(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {},
			"bob-token":   {},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap alice
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("X-Adapter-Actor", "discord:test-server:alice-user")
	server.ServeHTTP(bootRec, req)
	if bootRec.Code != http.StatusOK {
		t.Fatalf("bootstrap alice: %d", bootRec.Code)
	}

	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community Comment")
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

	// Post comment on jump
	commentBody := map[string]string{"body": "Great jump!"}
	commentRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/comments", "bob-token", commentBody)
	if commentRec.Code != http.StatusCreated {
		t.Fatalf("post jump comment: expected 201, got %d: %s", commentRec.Code, commentRec.Body.String())
	}

	type commentDTO struct {
		ID       string `json:"id"`
		Body     string `json:"body"`
		JumpID   string `json:"jumpId"`
		PlayerID string `json:"playerId"`
	}
	type commentResp struct {
		Comment commentDTO `json:"comment"`
	}
	var cr commentResp
	decodeResponse(t, commentRec, &cr)
	if cr.Comment.Body != "Great jump!" {
		t.Fatalf("expected Body 'Great jump!', got %q", cr.Comment.Body)
	}
	if cr.Comment.JumpID != subResp.Jump.ID {
		t.Fatalf("expected JumpID %s, got %s", subResp.Jump.ID, cr.Comment.JumpID)
	}
}

func TestPostCommentOnRevealedRound(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

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

	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community Round Comment")
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

	// Reveal
	store.SetClock(func() time.Time { return past.Add(25 * time.Hour) })
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("reveal: %d", revealRec.Code)
	}

	// Post round-level comment
	commentBody := map[string]string{"body": "What a round!"}
	commentRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/comments", "alice-token", commentBody)
	if commentRec.Code != http.StatusCreated {
		t.Fatalf("post round comment: expected 201, got %d: %s", commentRec.Code, commentRec.Body.String())
	}

	type commentDTO struct {
		ID     string `json:"id"`
		Body   string `json:"body"`
		JumpID string `json:"jumpId"`
	}
	type commentResp struct {
		Comment commentDTO `json:"comment"`
	}
	var cr commentResp
	decodeResponse(t, commentRec, &cr)
	if cr.Comment.Body != "What a round!" {
		t.Fatalf("expected Body 'What a round!', got %q", cr.Comment.Body)
	}
	if cr.Comment.JumpID != "" {
		t.Fatalf("expected empty JumpID for round-level comment, got %q", cr.Comment.JumpID)
	}
}

func TestPostCommentFailsOnNonRevealedRound(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time { return time.Now().UTC() })

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

	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community NoReveal Comment")
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

	// Find longest timeframe
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

	// Try comment on non-revealed round
	commentBody := map[string]string{"body": "Too early!"}
	commentRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/comments", "alice-token", commentBody)
	if commentRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on non-revealed round, got %d: %s", commentRec.Code, commentRec.Body.String())
	}
}

func TestPostCommentRequiresAuth(t *testing.T) {
	server := newTestServer(t)

	commentBody := map[string]string{"body": "test"}
	rec := doJSON(server, http.MethodPost, "/v1/rounds/round-1/comments", "", commentBody)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPostCommentFailsOnEmptyBody(t *testing.T) {
	server := newTestServer(t)

	commentBody := map[string]string{}
	rec := doJSON(server, http.MethodPost, "/v1/rounds/round-1/comments", "alice-token", commentBody)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on empty body, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListCommentsForJump(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {},
			"bob-token":   {},
		},
		Store: store,
		Now:   store.Now,
	})

	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("X-Adapter-Actor", "discord:test-server:alice-user")
	server.ServeHTTP(bootRec, req)

	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community List Comment")
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
		"caption":      "List test jump",
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
		t.Fatalf("reveal: %d", revealRec.Code)
	}

	// Post comments
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/comments", "alice-token", map[string]string{"body": "First!"})
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/comments", "bob-token", map[string]string{"body": "Second!"})
	// Post a round-level comment too
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/comments", "alice-token", map[string]string{"body": "Round chat"})

	// List jump comments
	listRec := doJSON(server, http.MethodGet, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/comments", "alice-token", nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list jump comments: expected 200, got %d: %s", listRec.Code, listRec.Body.String())
	}

	type listResp struct {
		Comments []struct {
			ID   string `json:"id"`
			Body string `json:"body"`
		} `json:"comments"`
	}
	var lr listResp
	decodeResponse(t, listRec, &lr)
	if len(lr.Comments) != 2 {
		t.Fatalf("expected 2 jump comments, got %d", len(lr.Comments))
	}

	// List round-level comments
	roundListRec := doJSON(server, http.MethodGet, "/v1/rounds/"+sr.Round.ID+"/comments", "alice-token", nil)
	if roundListRec.Code != http.StatusOK {
		t.Fatalf("list round comments: expected 200, got %d: %s", roundListRec.Code, roundListRec.Body.String())
	}
	var rlr listResp
	decodeResponse(t, roundListRec, &rlr)
	if len(rlr.Comments) != 1 {
		t.Fatalf("expected 1 round-level comment, got %d", len(rlr.Comments))
	}
	if rlr.Comments[0].Body != "Round chat" {
		t.Fatalf("expected 'Round chat', got %q", rlr.Comments[0].Body)
	}
}

func TestNonJumperCanComment(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {},
			"bob-token":   {},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap both
	bootRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer alice-token")
	req.Header.Set("X-Adapter-Actor", "discord:test-server:alice-user")
	server.ServeHTTP(bootRec, req)

	bootRec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req2.Header.Set("Authorization", "Bearer bob-token")
	req2.Header.Set("X-Adapter-Actor", "discord:test-server:bob-user")
	server.ServeHTTP(bootRec2, req2)

	result, err := store.ResolveExternalActor(context.Background(), "discord", "test-server", "alice-user", "Alice", "Test Community NonJumper")
	if err != nil {
		t.Fatalf("resolve alice: %v", err)
	}
	_, err = store.ResolveExternalActor(context.Background(), "discord", "test-server", "bob-user", "Bob", result.CommunityID)
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

	// Only alice commits and submits
	commitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "alice-token", nil)
	if commitRec.Code != http.StatusCreated {
		t.Fatalf("commit: %d", commitRec.Code)
	}

	jumpBody := map[string]any{
		"caption":      "Alice's jump",
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
		t.Fatalf("reveal: %d", revealRec.Code)
	}

	// Bob (non-jumper) comments on alice's jump
	commentBody := map[string]string{"body": "Nice work Alice!"}
	commentRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/comments", "bob-token", commentBody)
	if commentRec.Code != http.StatusCreated {
		t.Fatalf("non-jumper comment: expected 201, got %d: %s", commentRec.Code, commentRec.Body.String())
	}

	type commentDTO struct {
		Body     string `json:"body"`
		PlayerID string `json:"playerId"`
	}
	type commentResp struct {
		Comment commentDTO `json:"comment"`
	}
	var cr commentResp
	decodeResponse(t, commentRec, &cr)
	if cr.Comment.Body != "Nice work Alice!" {
		t.Fatalf("expected Body 'Nice work Alice!', got %q", cr.Comment.Body)
	}
}
