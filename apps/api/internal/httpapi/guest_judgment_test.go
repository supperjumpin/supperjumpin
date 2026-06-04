package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func doJSONUnauthenticated(server http.Handler, method string, path string, body any) *httptest.ResponseRecorder {
	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			panic(err)
		}
	}
	req := httptest.NewRequest(method, path, &requestBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec
}

func testGuestCanCreateSession(t *testing.T) {
	server := newTestServer(t)

	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/guest-sessions", nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode guest session response: %v", err)
	}
	if body["id"] == "" {
		t.Fatal("expected guest session id in response")
	}
}

func testGuestCanJudgePerformedJump(t *testing.T) {
	server, store := newTestServerAndStore(t)

	// Create a guest session
	sessionRec := doJSON(server, http.MethodPost, "/v1/guest-sessions", "", nil)
	if sessionRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for guest session, got %d: %s", sessionRec.Code, sessionRec.Body.String())
	}
	var sessionBody map[string]string
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode guest session: %v", err)
	}
	guestSessionID := sessionBody["id"]

	// Create a performed jump by Alice
	performed := performJump(t, server, "alice-token")

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	// Guest judges the jump
	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.ID+"/judgment", map[string]any{
		"guestSessionId": guestSessionID,
		"commitment":     2,
		"transgression":  3,
		"creativity":     3,
		"presentation":   4,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var judgment judgmentBody
	decodeResponse(t, rec, &judgment)
	if judgment.JumpID != performed.ID {
		t.Fatalf("expected judgment for jump %s, got %s", performed.ID, judgment.JumpID)
	}
	if judgment.PlayerID != "" {
		t.Fatalf("expected empty player id for guest judgment, got %q", judgment.PlayerID)
	}
}

func testGuestEditDoesNotIncrementCap(t *testing.T) {
	server, store := newTestServerAndStore(t)

	sessionRec := doJSON(server, http.MethodPost, "/v1/guest-sessions", "", nil)
	var sessionBody map[string]string
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode guest session: %v", err)
	}
	guestSessionID := sessionBody["id"]

	store.SetClock(time.Now)
	performed := performJump(t, server, "alice-token")
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	first := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.ID+"/judgment", map[string]any{
		"guestSessionId": guestSessionID,
		"commitment":     2,
		"transgression":  3,
		"creativity":     3,
		"presentation":   4,
	})
	if first.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for first judgment, got %d: %s", first.Code, first.Body.String())
	}

	countAfterFirst, err := store.GuestSessionJudgmentCount(context.Background(), guestSessionID)
	if err != nil {
		t.Fatalf("count guest judgments after first submission: %v", err)
	}
	if countAfterFirst != 1 {
		t.Fatalf("expected guest cap count 1 after first submission, got %d", countAfterFirst)
	}

	edit := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.ID+"/judgment", map[string]any{
		"guestSessionId": guestSessionID,
		"commitment":     4,
		"transgression":  4,
		"creativity":     4,
		"presentation":   4,
	})
	if edit.Code != http.StatusOK {
		t.Fatalf("expected status 200 for edit, got %d: %s", edit.Code, edit.Body.String())
	}

	countAfterEdit, err := store.GuestSessionJudgmentCount(context.Background(), guestSessionID)
	if err != nil {
		t.Fatalf("count guest judgments after edit: %v", err)
	}
	if countAfterEdit != 1 {
		t.Fatalf("expected guest cap count to stay 1 after edit, got %d", countAfterEdit)
	}
}

func testGuestConcurrentJudgmentsAtCapBoundary(t *testing.T) {
	server, store := newTestServerAndStore(t)

	sessionRec := doJSON(server, http.MethodPost, "/v1/guest-sessions", "", nil)
	var sessionBody map[string]string
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode guest session: %v", err)
	}
	guestSessionID := sessionBody["id"]

	for i := 0; i < 4; i++ {
		store.SetClock(time.Now)
		performed := performJump(t, server, "alice-token")
		store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })
		rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.ID+"/judgment", map[string]any{
			"guestSessionId": guestSessionID,
			"commitment":     2,
			"transgression":  2,
			"creativity":     3,
			"presentation":   3,
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 for seed judgment %d, got %d: %s", i+1, rec.Code, rec.Body.String())
		}
	}

	countBefore, err := store.GuestSessionJudgmentCount(context.Background(), guestSessionID)
	if err != nil {
		t.Fatalf("count guest judgments before concurrent submissions: %v", err)
	}
	if countBefore != 4 {
		t.Fatalf("expected guest cap count 4 before concurrent submissions, got %d", countBefore)
	}

	store.SetClock(time.Now)
	firstJump := performJump(t, server, "alice-token")
	secondJump := performJump(t, server, "alice-token")
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	start := make(chan struct{})
	results := make(chan int, 2)
	var wg sync.WaitGroup
	judge := func(jumpID string) {
		defer wg.Done()
		<-start
		rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+jumpID+"/judgment", map[string]any{
			"guestSessionId": guestSessionID,
			"commitment":     2,
			"transgression":  2,
			"creativity":     3,
			"presentation":   3,
		})
		results <- rec.Code
	}

	wg.Add(2)
	go judge(firstJump.ID)
	go judge(secondJump.ID)
	close(start)
	wg.Wait()
	close(results)

	codes := make([]int, 0, 2)
	for code := range results {
		codes = append(codes, code)
	}
	if len(codes) != 2 {
		t.Fatalf("expected 2 concurrent responses, got %d", len(codes))
	}
	var created, forbidden int
	for _, code := range codes {
		switch code {
		case http.StatusCreated:
			created++
		case http.StatusForbidden:
			forbidden++
		default:
			t.Fatalf("expected one 201 and one 403, got codes %v", codes)
		}
	}
	if created != 1 || forbidden != 1 {
		t.Fatalf("expected one 201 and one 403, got codes %v", codes)
	}

	countAfter, err := store.GuestSessionJudgmentCount(context.Background(), guestSessionID)
	if err != nil {
		t.Fatalf("count guest judgments after concurrent submissions: %v", err)
	}
	if countAfter != 5 {
		t.Fatalf("expected guest cap count 5 after concurrent submissions, got %d", countAfter)
	}
}

func testGuestCapBlocksAdditionalJudgments(t *testing.T) {
	server, store := newTestServerAndStore(t)

	// Create a guest session
	sessionRec := doJSON(server, http.MethodPost, "/v1/guest-sessions", "", nil)
	var sessionBody map[string]string
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode guest session: %v", err)
	}
	guestSessionID := sessionBody["id"]

	// Guest submits 5 judgments (cap)
	for i := 0; i < 5; i++ {
		// Reset clock so performJump creates a consistent grace period, then advance past it
		store.SetClock(time.Now)
		performed := performJump(t, server, "alice-token")

		// Advance past the Author Grace Period before judging
		store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

		rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.ID+"/judgment", map[string]any{
			"guestSessionId": guestSessionID,
			"commitment":     2,
			"transgression":  2,
			"creativity":     3,
			"presentation":   3,
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 for judgment %d, got %d: %s", i+1, rec.Code, rec.Body.String())
		}
	}

	// 6th judgment should be blocked
	performed := performJump(t, server, "alice-token")
	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.ID+"/judgment", map[string]any{
		"guestSessionId": guestSessionID,
		"commitment":     2,
		"transgression":  2,
		"creativity":     3,
		"presentation":   3,
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for cap-exceeded judgment, got %d: %s", rec.Code, rec.Body.String())
	}
}

func testAuthenticatedPlayerJudgmentStillWorks(t *testing.T) {
	server, store := newTestServerAndStore(t)

	performed := performJump(t, server, "alice-token")

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	// Bob (authenticated) judges Alice's jump
	judgment := submitJudgment(t, server, "bob-token", performed.ID, 2, 3, 3, 4, http.StatusCreated)
	if judgment.JumpID != performed.ID {
		t.Fatalf("expected judgment for jump %s, got %s", performed.ID, judgment.JumpID)
	}
	if judgment.PlayerID == "" {
		t.Fatal("expected player id on authenticated judgment")
	}
}

func testCannotProvideBothAuthAndGuestSession(t *testing.T) {
	server := newTestServer(t)

	performed := performJump(t, server, "alice-token")

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.ID+"/judgment", "bob-token", map[string]any{
		"guestSessionId": "some-session-id",
		"commitment":     2,
		"transgression":  2,
		"creativity":     3,
		"presentation":   3,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 when both auth and guestSessionId provided, got %d: %s", rec.Code, rec.Body.String())
	}
}

func testGuestSessionIdRequiredWhenUnauthenticated(t *testing.T) {
	server := newTestServer(t)

	performed := performJump(t, server, "alice-token")

	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.ID+"/judgment", map[string]any{
		"commitment":    2,
		"transgression": 2,
		"creativity":    3,
		"presentation":  3,
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when unauthenticated without guestSessionId, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGuestJudgment(t *testing.T) {
	t.Run("guest sessions", func(t *testing.T) {
		t.Run("guest can create session", testGuestCanCreateSession)
	})

	t.Run("guest judgments", func(t *testing.T) {
		t.Run("guest can judge performed jump", testGuestCanJudgePerformedJump)
		t.Run("guest edit does not increment cap", testGuestEditDoesNotIncrementCap)
		t.Run("guest concurrent judgments at cap boundary", testGuestConcurrentJudgmentsAtCapBoundary)
		t.Run("guest cap blocks additional judgments", testGuestCapBlocksAdditionalJudgments)
	})

	t.Run("authenticated judgment", func(t *testing.T) {
		t.Run("authenticated player judgment still works", testAuthenticatedPlayerJudgmentStillWorks)
	})

	t.Run("request validation", func(t *testing.T) {
		t.Run("cannot provide both auth and guest session", testCannotProvideBothAuthAndGuestSession)
		t.Run("guest session id required when unauthenticated", testGuestSessionIdRequiredWhenUnauthenticated)
	})
}
