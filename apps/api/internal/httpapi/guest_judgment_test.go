package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	server := newGroupsTestServer(t)

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
	server, store := newGroupsTestServerAndStore(t)

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

	// Create a group and a performed jump by Alice
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	// Guest judges the jump
	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", map[string]any{
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
	if judgment.JumpID != performed.Jump.ID {
		t.Fatalf("expected judgment for jump %s, got %s", performed.Jump.ID, judgment.JumpID)
	}
	if judgment.PlayerID != "" {
		t.Fatalf("expected empty player id for guest judgment, got %q", judgment.PlayerID)
	}
}

func testGuestCapBlocksAdditionalJudgments(t *testing.T) {
	server, store := newGroupsTestServerAndStore(t)

	// Create a guest session
	sessionRec := doJSON(server, http.MethodPost, "/v1/guest-sessions", "", nil)
	var sessionBody map[string]string
	if err := json.NewDecoder(sessionRec.Body).Decode(&sessionBody); err != nil {
		t.Fatalf("decode guest session: %v", err)
	}
	guestSessionID := sessionBody["id"]

	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	// Guest submits 5 judgments (cap)
	for i := 0; i < 5; i++ {
		// Reset clock so performJump creates a consistent grace period, then advance past it
		store.SetClock(time.Now)
		performed := performJump(t, server, "alice-token", group.Group.ID)

		// Advance past the Author Grace Period before judging
		store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

		rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", map[string]any{
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
	performed := performJump(t, server, "alice-token", group.Group.ID)
	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", map[string]any{
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
	server, store := newGroupsTestServerAndStore(t)

	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	// Bob (authenticated) judges Alice's jump
	judgment := submitJudgment(t, server, "bob-token", performed.Jump.ID, 2, 3, 3, 4, http.StatusCreated)
	if judgment.JumpID != performed.Jump.ID {
		t.Fatalf("expected judgment for jump %s, got %s", performed.Jump.ID, judgment.JumpID)
	}
	if judgment.PlayerID == "" {
		t.Fatal("expected player id on authenticated judgment")
	}
}

func testCannotProvideBothAuthAndGuestSession(t *testing.T) {
	server := newGroupsTestServer(t)

	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]any{
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
	server := newGroupsTestServer(t)

	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	rec := doJSONUnauthenticated(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", map[string]any{
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
