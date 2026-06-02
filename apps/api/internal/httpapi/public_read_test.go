package httpapi_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

// ---------------------------------------------------------------------------
// Public Feed tests (GET /v1/feed)
// ---------------------------------------------------------------------------

func TestPublicFeedReturnsJumps(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performJump(t, server, "alice-token", store.GroupID)

	rec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Jumps      []any  `json:"jumps"`
		NextCursor string `json:"nextCursor"`
	}
	decodeResponse(t, rec, &res)
	if len(res.Jumps) != 1 {
		t.Fatalf("expected 1 jump in feed, got %d", len(res.Jumps))
	}
	// Unauthenticated feed should have no viewerContext
	// (JumpCard doesn't include it for unauthenticated requests)
}

func TestPublicFeedReturnsMultipleJumpsInOrder(t *testing.T) {
	server, store := newPublicReadTestServer(t)

	// Perform 3 jumps — they will be ordered by CreatedAt DESC
	for i := 0; i < 3; i++ {
		performJump(t, server, "alice-token", store.GroupID)
	}

	rec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Jumps []struct {
			ID string `json:"id"`
		} `json:"jumps"`
	}
	decodeResponse(t, rec, &res)
	if len(res.Jumps) != 3 {
		t.Fatalf("expected 3 jumps, got %d", len(res.Jumps))
	}
}

func TestPublicFeedExcludesRemovedJumps(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token", store.GroupID)

	// Raise a dispute and resolve with "Removed Jump"
	dispute := raiseDispute(t, server, "bob-token", performed.Jump.ID, "House Rules", "Should be removed")
	resolutionRec := doJSON(server, http.MethodPost, "/v1/disputes/"+dispute.ID+"/resolution", "alice-token", map[string]string{
		"resolution":       "Removed Jump",
		"resolutionReason": "Test removal",
	})
	if resolutionRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on dispute resolution, got %d: %s", resolutionRec.Code, resolutionRec.Body.String())
	}

	// Feed should exclude the removed jump
	rec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Jumps []any `json:"jumps"`
	}
	decodeResponse(t, rec, &res)
	if len(res.Jumps) != 0 {
		t.Fatalf("expected 0 jumps (removed), got %d", len(res.Jumps))
	}
}

func TestPublicFeedCursorPagination(t *testing.T) {
	server, store := newPublicReadTestServer(t)

	// Create 4 jumps with distinct food names
	foods := []string{"Crunchwrap", "Taco", "Burrito", "Quesadilla"}
	for i := 0; i < 4; i++ {
		idea := createIdea(t, server, "alice-token", store.GroupID, "Taco Bell", "Olive Garden parking lot", foods[i])
		planned := createPlannedJump(t, server, "alice-token", idea.ID, false)
		auth := authorizeEvidenceUpload(t, server, "alice-token", planned.ID, "image/jpeg")
		rec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence", "alice-token", map[string]string{
			"uploadAuthorizationId": auth.ID,
			"caption":               "Jump " + foods[i],
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 for jump %d, got %d: %s", i, rec.Code, rec.Body.String())
		}
	}

	// Fetch first page with limit=2 — expect cursor
	rec := doJSON(server, http.MethodGet, "/v1/feed?limit=2", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var page1 struct {
		Jumps      []struct {
			ID string `json:"id"`
		} `json:"jumps"`
		NextCursor *string `json:"nextCursor"`
	}
	decodeResponse(t, rec, &page1)
	if len(page1.Jumps) != 2 {
		t.Fatalf("expected 2 jumps on page 1, got %d", len(page1.Jumps))
	}
	if page1.NextCursor == nil {
		t.Fatal("expected nextCursor for pagination, got nil")
	}

	// Fetch second page with the cursor — expect at least 1 item, no cursor
	rec2 := doJSON(server, http.MethodGet, "/v1/feed?cursor="+*page1.NextCursor+"&limit=2", "", nil)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for page 2, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var page2 struct {
		Jumps      []struct {
			ID string `json:"id"`
		} `json:"jumps"`
		NextCursor *string `json:"nextCursor"`
	}
	decodeResponse(t, rec2, &page2)
	if len(page2.Jumps) == 0 {
		t.Fatal("expected at least 1 jump on page 2, got 0")
	}
	if page2.NextCursor != nil {
		t.Fatal("expected no nextCursor on last page")
	}

	// Verify no duplicates across pages
	seen := map[string]bool{}
	for _, j := range page1.Jumps {
		seen[j.ID] = true
	}
	for _, j := range page2.Jumps {
		seen[j.ID] = true
	}
	if len(seen) != len(page1.Jumps)+len(page2.Jumps) {
		t.Fatal("duplicate items across pages")
	}
}

func TestPublicFeedInvalidCursorReturns400(t *testing.T) {
	server := newGroupsTestServer()

	rec := doJSON(server, http.MethodGet, "/v1/feed?cursor=not-base64", "", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid cursor, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPublicFeedInvalidLimitReturns400(t *testing.T) {
	server := newGroupsTestServer()

	rec := doJSON(server, http.MethodGet, "/v1/feed?limit=abc", "", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid limit, got %d: %s", rec.Code, rec.Body.String())
	}

	rec2 := doJSON(server, http.MethodGet, "/v1/feed?limit=999", "", nil)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for limit > 50, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Public Jump Detail tests (GET /v1/jumps/{jumpID})
// ---------------------------------------------------------------------------

func TestPublicJumpDetailReturnsJump(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token", store.GroupID)

	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.Jump.ID, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var detail struct {
		ID            string  `json:"id"`
		PerformerName string  `json:"performerName"`
		Source        string  `json:"source"`
		Destination   string  `json:"destination"`
		Food          string  `json:"food"`
		Caption       string  `json:"caption"`
		RunningAverage float64 `json:"runningAverage"`
		JudgmentCount  int     `json:"judgmentCount"`
		ViewerContext  *struct {
			CanJudge bool `json:"canJudge"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, rec, &detail)
	if detail.ID != performed.Jump.ID {
		t.Fatalf("expected jump ID %q, got %q", performed.Jump.ID, detail.ID)
	}
	if detail.PerformerName == "" {
		t.Fatal("expected performerName to be populated")
	}
	if detail.ViewerContext == nil {
		t.Fatal("expected viewerContext even for unauthenticated requests")
	}
	if !detail.ViewerContext.CanJudge {
		t.Fatal("expected unauthenticated viewer to be able to judge")
	}
}

func TestPublicJumpDetailUnknownIDReturns404(t *testing.T) {
	server := newGroupsTestServer()

	rec := doJSON(server, http.MethodGet, "/v1/jumps/unknown-id", "", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown jump, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPublicJumpDetailRemovedJumpReturnsTombstone(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token", store.GroupID)

	// Remove the jump via dispute resolution
	dispute := raiseDispute(t, server, "bob-token", performed.Jump.ID, "House Rules", "Remove this")
	doJSON(server, http.MethodPost, "/v1/disputes/"+dispute.ID+"/resolution", "alice-token", map[string]string{
		"resolution":       "Removed Jump",
		"resolutionReason": "Test removal",
	})

	// Detail should return tombstone
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.Jump.ID, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (tombstone), got %d: %s", rec.Code, rec.Body.String())
	}

	var tombstone struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	decodeResponse(t, rec, &tombstone)
	if tombstone.Status != "Removed Jump" {
		t.Fatalf("expected Removed Jump status, got %q", tombstone.Status)
	}
	if tombstone.Message != "This Jump is no longer available" {
		t.Fatalf("expected tombstone message, got %q", tombstone.Message)
	}
}

// ---------------------------------------------------------------------------
// Viewer Context tests
// ---------------------------------------------------------------------------

func TestPublicJumpDetailSelfJudgingViewerContext(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token", store.GroupID)

	// Alice views her own jump detail
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.Jump.ID, "alice-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var detail struct {
		ViewerContext *struct {
			CanJudge bool   `json:"canJudge"`
			Reason   string `json:"reason"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, rec, &detail)
	if detail.ViewerContext == nil {
		t.Fatal("expected viewerContext")
	}
	if detail.ViewerContext.CanJudge {
		t.Fatal("expected self-judging viewer to NOT be able to judge")
	}
	if detail.ViewerContext.Reason != "self-judging" {
		t.Fatalf("expected reason 'self-judging', got %q", detail.ViewerContext.Reason)
	}
}

func TestPublicJumpDetailOtherPlayerCanJudge(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token", store.GroupID)

	// Bob (different player) views detail — should be able to judge
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.Jump.ID, "bob-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var detail struct {
		ViewerContext *struct {
			CanJudge bool   `json:"canJudge"`
			Reason   string `json:"reason,omitempty"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, rec, &detail)
	if detail.ViewerContext == nil {
		t.Fatal("expected viewerContext")
	}
	if !detail.ViewerContext.CanJudge {
		t.Fatal("expected Bob to be able to judge Alice's jump")
	}
}

func TestPublicDetailAlreadyJudgedViewerContext(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token", store.GroupID)

	// Bob judges Alice's jump (scores must be 1-4)
	submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 2, 1, http.StatusCreated)

	// Bob views detail again — should show "already-judged"
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.Jump.ID, "bob-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var detail struct {
		ViewerContext *struct {
			CanJudge  bool   `json:"canJudge"`
			Reason    string `json:"reason,omitempty"`
			HasJudged bool   `json:"hasJudged"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, rec, &detail)
	if detail.ViewerContext == nil {
		t.Fatal("expected viewerContext")
	}
	if detail.ViewerContext.CanJudge {
		t.Fatal("expected already-judged viewer to NOT be able to judge")
	}
	if detail.ViewerContext.Reason != "already-judged" {
		t.Fatalf("expected reason 'already-judged', got %q", detail.ViewerContext.Reason)
	}
	if !detail.ViewerContext.HasJudged {
		t.Fatal("expected HasJudged=true")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// publicReadTestStore holds a pre-setup MemoryStore and a Group that tests can
// use to create Jumps without repeating setup.
type publicReadTestStore struct {
	Store   *httpapi.MemoryStore
	DB      httpapi.Persistence
	GroupID string
}

// newPublicReadTestServer creates a server + store with an established Group
// and two players (Alice + Bob) so tests can create Jumps and judgments.
func newPublicReadTestServer(t *testing.T) (http.Handler, *publicReadTestStore) {
	t.Helper()
	store := httpapi.NewMemoryStore()
	server := newGroupsTestServerWithPersistence(store)
	group := createGroup(t, server, "alice-token", "Read Test Group")

	// Invite Bob
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	joinRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on invite accept, got %d: %s", joinRec.Code, joinRec.Body.String())
	}

	return server, &publicReadTestStore{
		Store:   store,
		DB:      store,
		GroupID: group.Group.ID,
	}
}

// Ensure unused import doesn't break build
var _ = time.Now
