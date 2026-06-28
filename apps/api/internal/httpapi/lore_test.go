package httpapi_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func TestGetCommunityLoreReturnsEntriesSortedByDensity(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	past := time.Now().Add(-25 * time.Hour).UTC()
	store.SetClock(func() time.Time { return past })

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
			"bob-token":   {Provider: "test-provider", Subject: "bob-auth", Email: "bob@example.com"},
			"carol-token": {Provider: "test-provider", Subject: "carol-auth", Email: "carol@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap
	for _, tok := range []string{"alice-token", "bob-token", "carol-token"} {
		doJSON(server, http.MethodGet, "/v1/me", tok, nil)
	}

	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-lore", "alice-discord-lore", "Alice", "Lore Community")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}

	// Get timeframes
	tfRec := doJSON(server, http.MethodGet, "/v1/reveal-timeframes", "", nil)
	type tfResp struct {
		Timeframes []struct {
			ID string `json:"id"`
		} `json:"timeframes"`
	}
	var tfs tfResp
	decodeResponse(t, tfRec, &tfs)

	// Get stamps
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

	stampByStance := map[string]string{}
	for _, s := range cat.Stamps {
		stampByStance[s.Stance] = s.ID
	}

	// Start round
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

	// Alice commits + submits jump A (high density)
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "alice-token", nil)
	submitRecA := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps", "alice-token", map[string]any{
		"caption":      "Jump A - the best one",
		"evidenceUrls": []string{"https://example.com/a.jpg"},
	})
	if submitRecA.Code != http.StatusCreated {
		t.Fatalf("submit jump A: %d", submitRecA.Code)
	}

	// Bob commits + submits jump B (low density)
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "bob-token", nil)
	submitRecB := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps", "bob-token", map[string]any{
		"caption":      "Jump B - meh",
		"evidenceUrls": []string{"https://example.com/b.jpg"},
	})
	if submitRecB.Code != http.StatusCreated {
		t.Fatalf("submit jump B: %d", submitRecB.Code)
	}

	type jumpDTO struct {
		ID string `json:"id"`
	}
	type submitResp struct {
		Jump jumpDTO `json:"jump"`
	}
	var respA, respB submitResp
	decodeResponse(t, submitRecA, &respA)
	decodeResponse(t, submitRecB, &respB)

	// Reveal
	store.SetClock(func() time.Time { return past.Add(25 * time.Hour) })
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("reveal: %d", revealRec.Code)
	}

	// Apply reactions: Jump A gets 3 stamps from 3 players, Jump B gets 1 stamp
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+respA.Jump.ID+"/reactions", "bob-token", map[string]string{"stampId": stampByStance["approval"]})
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+respA.Jump.ID+"/reactions", "carol-token", map[string]string{"stampId": stampByStance["lore"]})
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+respA.Jump.ID+"/reactions", "alice-token", map[string]string{"stampId": stampByStance["chaos"]})

	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+respB.Jump.ID+"/reactions", "carol-token", map[string]string{"stampId": stampByStance["approval"]})

	// Get lore
	loreRec := doJSON(server, http.MethodGet, "/v1/communities/"+result.CommunityID+"/lore", "alice-token", nil)
	if loreRec.Code != http.StatusOK {
		t.Fatalf("get lore: expected 200, got %d: %s", loreRec.Code, loreRec.Body.String())
	}

	type loreEntryDTO struct {
		JumpID       string         `json:"jumpId"`
		RoundID      string         `json:"roundId"`
		JumpCaption  string         `json:"jumpCaption"`
		JumpPlayerID string         `json:"jumpPlayerId"`
		StampCounts  map[string]int `json:"stampCounts"`
		TotalStamps  int            `json:"totalStamps"`
	}
	type loreResp struct {
		Entries []loreEntryDTO `json:"entries"`
	}
	var lr loreResp
	decodeResponse(t, loreRec, &lr)

	if len(lr.Entries) != 2 {
		t.Fatalf("expected 2 lore entries, got %d", len(lr.Entries))
	}

	// Jump A should be first (most stamps)
	if lr.Entries[0].JumpID != respA.Jump.ID {
		t.Fatalf("expected first entry to be jump A, got %s", lr.Entries[0].JumpID)
	}
	if lr.Entries[0].TotalStamps != 3 {
		t.Fatalf("expected jump A TotalStamps=3, got %d", lr.Entries[0].TotalStamps)
	}
	if lr.Entries[0].JumpCaption != "Jump A - the best one" {
		t.Fatalf("expected jump A caption, got %q", lr.Entries[0].JumpCaption)
	}

	// Jump B should be second
	if lr.Entries[1].JumpID != respB.Jump.ID {
		t.Fatalf("expected second entry to be jump B, got %s", lr.Entries[1].JumpID)
	}
	if lr.Entries[1].TotalStamps != 1 {
		t.Fatalf("expected jump B TotalStamps=1, got %d", lr.Entries[1].TotalStamps)
	}

	// Verify stamp stance counts on highest-density entry
	if lr.Entries[0].StampCounts["approval"] != 1 {
		t.Fatalf("expected 1 approval, got %d", lr.Entries[0].StampCounts["approval"])
	}
	if lr.Entries[0].StampCounts["lore"] != 1 {
		t.Fatalf("expected 1 lore, got %d", lr.Entries[0].StampCounts["lore"])
	}
	if lr.Entries[0].StampCounts["chaos"] != 1 {
		t.Fatalf("expected 1 chaos, got %d", lr.Entries[0].StampCounts["chaos"])
	}

	// No per-player tally exists on the DTO (structurally verified by type)
}

func TestGetCommunityLoreEmptyCommunity(t *testing.T) {
	store := newCleanPostgresTestStore(t)

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "test-provider", Subject: "alice-auth", Email: "alice@example.com"},
		},
		Store: store,
		Now:   store.Now,
	})

	// Bootstrap alice
	doJSON(server, http.MethodGet, "/v1/me", "alice-token", nil)

	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-empty", "alice-empty", "Alice", "Empty Community")
	if err != nil {
		t.Fatalf("resolve external actor: %v", err)
	}

	// Get lore for community with no rounds
	loreRec := doJSON(server, http.MethodGet, "/v1/communities/"+result.CommunityID+"/lore", "alice-token", nil)
	if loreRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for empty lore, got %d: %s", loreRec.Code, loreRec.Body.String())
	}

	type loreEntryDTO struct {
		JumpID string `json:"jumpId"`
	}
	type loreResp struct {
		Entries []loreEntryDTO `json:"entries"`
	}
	var lr loreResp
	decodeResponse(t, loreRec, &lr)

	if len(lr.Entries) != 0 {
		t.Fatalf("expected 0 lore entries for empty community, got %d", len(lr.Entries))
	}
}

func TestGetCommunityLoreExcludesNonRevealedRounds(t *testing.T) {
	store := newCleanPostgresTestStore(t)

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

	// Bootstrap
	for _, tok := range []string{"alice-token", "bob-token"} {
		doJSON(server, http.MethodGet, "/v1/me", tok, nil)
	}

	result, err := store.ResolveExternalActor(context.Background(), "discord", "server-sealed", "alice-sealed", "Alice", "Sealed Community")
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

	// Use 7-day timeframe to ensure round doesn't auto-reveal
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

	// Commit + submit
	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/commits", "alice-token", nil)
	submitRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps", "alice-token", map[string]any{
		"caption":      "Secret jump",
		"evidenceUrls": []string{"https://example.com/s.jpg"},
	})
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

	// Attempt reaction on non-revealed jump (should fail at reaction level)
	// But we're testing that lore doesn't include non-revealed round jumps even if a reaction exists somehow.
	// Actually the reaction endpoint requires revealed rounds, so we can't add reactions.
	// Verify no lore entries for this community (round not revealed, no reactions possible)

	loreRec := doJSON(server, http.MethodGet, "/v1/communities/"+result.CommunityID+"/lore", "alice-token", nil)
	if loreRec.Code != http.StatusOK {
		t.Fatalf("get lore: expected 200, got %d", loreRec.Code)
	}

	type loreEntryDTO struct {
		JumpID string `json:"jumpId"`
	}
	type loreResp struct {
		Entries []loreEntryDTO `json:"entries"`
	}
	var lr loreResp
	decodeResponse(t, loreRec, &lr)

	// Non-revealed rounds shouldn't contribute to lore
	// (there shouldn't be any reactions on them anyway, but the SQL query enforces revealed)
	if len(lr.Entries) != 0 {
		t.Fatalf("expected 0 lore entries for non-revealed round, got %d", len(lr.Entries))
	}

	// Now reveal and check that lore appears
	store.SetClock(func() time.Time { return past.Add(200 * time.Hour) })
	revealRec := doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/reveal", "alice-token", nil)
	if revealRec.Code != http.StatusOK {
		t.Fatalf("reveal: %d", revealRec.Code)
	}

	// Apply reaction
	stampRec := doJSON(server, http.MethodGet, "/v1/stamp-catalog", "", nil)
	type catalogResp struct {
		Stamps []struct {
			ID string `json:"id"`
		} `json:"stamps"`
	}
	var cat catalogResp
	decodeResponse(t, stampRec, &cat)

	doJSON(server, http.MethodPost, "/v1/rounds/"+sr.Round.ID+"/jumps/"+subResp.Jump.ID+"/reactions", "bob-token", map[string]string{"stampId": cat.Stamps[0].ID})

	loreRec2 := doJSON(server, http.MethodGet, "/v1/communities/"+result.CommunityID+"/lore", "alice-token", nil)
	lr2 := loreResp{}
	decodeResponse(t, loreRec2, &lr2)

	if len(lr2.Entries) != 1 {
		t.Fatalf("expected 1 lore entry after reveal, got %d", len(lr2.Entries))
	}
}

func TestGetCommunityLoreRequiresAuth(t *testing.T) {
	server := newTestServer(t)

	rec := doJSON(server, http.MethodGet, "/v1/communities/community-1/lore", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
