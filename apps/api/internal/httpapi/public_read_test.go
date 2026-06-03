package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

// ---------------------------------------------------------------------------
// Public Feed tests (GET /v1/feed)
// ---------------------------------------------------------------------------

func testPublicFeedReturnsJumps(t *testing.T) {
	server, _ := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token")

	rec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res publicFeedBody
	decodeResponse(t, rec, &res)
	if len(res.Jumps) != 1 {
		t.Fatalf("expected 1 jump in feed, got %d", len(res.Jumps))
	}
	card := res.Jumps[0]
	if card.ID != performed.ID {
		t.Fatalf("expected jump ID %q, got %q", performed.ID, card.ID)
	}
	if card.PerformerName != "alice" {
		t.Fatalf("expected performerName alice, got %q", card.PerformerName)
	}
	if card.Source != performed.Source || card.Destination != performed.Destination || card.Food != performed.Food {
		t.Fatalf("expected route %q -> %q with food %q, got %#v", performed.Source, performed.Destination, performed.Food, card)
	}
	if card.Caption != "Crunchwrap successfully smuggled into the parking lot." {
		t.Fatalf("expected caption %q, got %q", "Crunchwrap successfully smuggled into the parking lot.", card.Caption)
	}
	if card.MediaObjectKey != "evidence_object_123" {
		t.Fatalf("expected mediaObjectKey %q, got %q", "evidence_object_123", card.MediaObjectKey)
	}
	if card.ViewerContext != nil {
		t.Fatal("expected unauthenticated feed card to omit viewerContext")
	}
	if res.NextCursor != nil {
		t.Fatalf("expected no nextCursor for single-card feed, got %q", *res.NextCursor)
	}
}

func testPublicFeedReturnsMultipleJumpsInOrder(t *testing.T) {
	server, store := newPublicReadTestServer(t)

	store.Store.SetClock(func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) })
	first := performJump(t, server, "alice-token")
	store.Store.SetClock(func() time.Time { return time.Date(2026, 6, 1, 12, 1, 0, 0, time.UTC) })
	second := performJump(t, server, "alice-token")
	store.Store.SetClock(func() time.Time { return time.Date(2026, 6, 1, 12, 2, 0, 0, time.UTC) })
	third := performJump(t, server, "alice-token")

	rec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res publicFeedBody
	decodeResponse(t, rec, &res)
	if len(res.Jumps) != 3 {
		t.Fatalf("expected 3 jumps, got %d", len(res.Jumps))
	}
	if res.Jumps[0].ID != third.ID || res.Jumps[1].ID != second.ID || res.Jumps[2].ID != first.ID {
		t.Fatalf("expected reverse chronological order [%q %q %q], got [%q %q %q]", third.ID, second.ID, first.ID, res.Jumps[0].ID, res.Jumps[1].ID, res.Jumps[2].ID)
	}
}

func testPublicFeedCursorPagination(t *testing.T) {
	server, store := newPublicReadTestServer(t)

	// Create 4 jumps with distinct timestamps.
	var created []string
	for i := 0; i < 4; i++ {
		minute := i
		store.Store.SetClock(func() time.Time {
			return time.Date(2026, 6, 1, 13, minute, 0, 0, time.UTC)
		})
		performed := performJump(t, server, "alice-token")
		created = append(created, performed.ID)
	}

	// Fetch first page with limit=2 — expect cursor
	rec := doJSON(server, http.MethodGet, "/v1/feed?limit=2", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var page1 publicFeedBody
	decodeResponse(t, rec, &page1)
	if len(page1.Jumps) != 2 {
		t.Fatalf("expected 2 jumps on page 1, got %d", len(page1.Jumps))
	}
	if page1.NextCursor == nil {
		t.Fatal("expected nextCursor for pagination, got nil")
	}
	if page1.Jumps[0].ID != created[3] || page1.Jumps[1].ID != created[2] {
		t.Fatalf("expected newest page-1 jumps [%q %q], got [%q %q]", created[3], created[2], page1.Jumps[0].ID, page1.Jumps[1].ID)
	}

	// Fetch second page with the cursor.
	rec2 := doJSON(server, http.MethodGet, "/v1/feed?cursor="+*page1.NextCursor+"&limit=2", "", nil)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for page 2, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var page2 publicFeedBody
	decodeResponse(t, rec2, &page2)
	if len(page2.Jumps) != 2 {
		t.Fatalf("expected 2 jumps on page 2, got %d", len(page2.Jumps))
	}
	if page2.NextCursor != nil {
		t.Fatal("expected no nextCursor on last page")
	}
	if page2.Jumps[0].ID != created[1] || page2.Jumps[1].ID != created[0] {
		t.Fatalf("expected older page-2 jumps [%q %q], got [%q %q]", created[1], created[0], page2.Jumps[0].ID, page2.Jumps[1].ID)
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

func testPublicFeedInvalidCursorReturns400(t *testing.T) {
	server := newTestServer(t)

	rec := doJSON(server, http.MethodGet, "/v1/feed?cursor=not-base64", "", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid cursor, got %d: %s", rec.Code, rec.Body.String())
	}
}

func testPublicFeedInvalidLimitReturns400(t *testing.T) {
	server := newTestServer(t)

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

func testPublicJumpDetailReturnsJump(t *testing.T) {
	server, _ := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token")

	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var detail struct {
		ID             string  `json:"id"`
		PerformerName  string  `json:"performerName"`
		Source         string  `json:"source"`
		Destination    string  `json:"destination"`
		Food           string  `json:"food"`
		Caption        string  `json:"caption"`
		MediaObjectKey string  `json:"mediaObjectKey"`
		RunningAverage float64 `json:"runningAverage"`
		JudgmentCount  int     `json:"judgmentCount"`
		ViewerContext  *struct {
			CanJudge bool `json:"canJudge"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, rec, &detail)
	if detail.ID != performed.ID {
		t.Fatalf("expected jump ID %q, got %q", performed.ID, detail.ID)
	}
	if detail.PerformerName == "" {
		t.Fatal("expected performerName to be populated")
	}
	if detail.Source != performed.Source || detail.Destination != performed.Destination || detail.Food != performed.Food {
		t.Fatalf("expected jump route %q -> %q with food %q, got %#v", performed.Source, performed.Destination, performed.Food, detail)
	}
	if detail.Caption != "Crunchwrap successfully smuggled into the parking lot." {
		t.Fatalf("expected caption %q, got %q", "Crunchwrap successfully smuggled into the parking lot.", detail.Caption)
	}
	if detail.MediaObjectKey != "evidence_object_123" {
		t.Fatalf("expected mediaObjectKey %q, got %q", "evidence_object_123", detail.MediaObjectKey)
	}
	if detail.ViewerContext == nil {
		t.Fatal("expected viewerContext even for unauthenticated requests")
	}
	if !detail.ViewerContext.CanJudge {
		t.Fatal("expected unauthenticated viewer to be able to judge")
	}
}

func testPublicJumpDetailUnknownIDReturns404(t *testing.T) {
	server := newTestServer(t)

	rec := doJSON(server, http.MethodGet, "/v1/jumps/unknown-id", "", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown jump, got %d: %s", rec.Code, rec.Body.String())
	}
	var missing struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	decodeResponse(t, rec, &missing)
	if missing.Error != "not_found" {
		t.Fatalf("expected not_found error code, got %q", missing.Error)
	}
	if missing.Message != "Jump not found. It may have been removed." {
		t.Fatalf("expected jump not found message, got %q", missing.Message)
	}
}

func testPublicFeedInternalErrorReturnsMessageEnvelope(t *testing.T) {
	store := &failingPublicReadStore{PostgresStore: newCleanPostgresTestStore(t)}
	store.SetClock(time.Now)
	server := newTestServerWithStore(store)

	rec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for feed load failure, got %d: %s", rec.Code, rec.Body.String())
	}

	var errBody struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	decodeResponse(t, rec, &errBody)
	if errBody.Error != "internal_error" {
		t.Fatalf("expected internal_error code, got %q", errBody.Error)
	}
	if errBody.Message != "Could not load jumps. Please try again." {
		t.Fatalf("expected public feed error message, got %q", errBody.Message)
	}
}

// ---------------------------------------------------------------------------
// Viewer Context tests
// ---------------------------------------------------------------------------

func testPublicJumpDetailSelfJudgingViewerContext(t *testing.T) {
	server, _ := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token")

	// Alice views her own jump detail
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "alice-token", nil)
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

func testPublicJumpDetailOtherPlayerCanJudge(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token")

	// Advance past the Author Grace Period so judging is available
	store.Store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	// Bob (different player) views detail — should be able to judge
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "bob-token", nil)
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

func testPublicDetailAlreadyJudgedViewerContext(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token")

	// Advance past the Author Grace Period before judging
	store.Store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	// Bob judges Alice's jump (scores must be 1-4)
	submitJudgment(t, server, "bob-token", performed.ID, 4, 3, 2, 1, http.StatusCreated)

	// Bob views detail again — should show "already-judged"
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "bob-token", nil)
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

func testPublicJumpDetailGracePeriodViewerContext(t *testing.T) {
	server, _ := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token")

	// Bob views Alice's jump while the Author Grace Period is still active
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "bob-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var detail struct {
		ViewerContext *struct {
			CanJudge          bool       `json:"canJudge"`
			Reason            string     `json:"reason,omitempty"`
			GracePeriodEndsAt *time.Time `json:"gracePeriodEndsAt,omitempty"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, rec, &detail)
	if detail.ViewerContext == nil {
		t.Fatal("expected viewerContext")
	}
	if detail.ViewerContext.CanJudge {
		t.Fatal("expected grace-period viewer to NOT be able to judge")
	}
	if detail.ViewerContext.Reason != "grace-period" {
		t.Fatalf("expected reason 'grace-period', got %q", detail.ViewerContext.Reason)
	}
	if detail.ViewerContext.GracePeriodEndsAt == nil {
		t.Fatal("expected gracePeriodEndsAt to be populated during grace period")
	}
}

func testPublicFeedGracePeriodViewerContext(t *testing.T) {
	server, _ := newPublicReadTestServer(t)
	performJump(t, server, "alice-token")

	// Bob requests the feed while the Author Grace Period is still active
	rec := doJSON(server, http.MethodGet, "/v1/feed", "bob-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res publicFeedBody
	decodeResponse(t, rec, &res)
	if len(res.Jumps) == 0 {
		t.Fatal("expected at least 1 jump in feed")
	}

	vc := res.Jumps[0].ViewerContext
	if vc == nil {
		t.Fatal("expected viewerContext on feed card for authenticated viewer")
	}
	if vc.CanJudge {
		t.Fatal("expected canJudge=false for grace-period jump on feed")
	}
	if vc.Reason != "grace-period" {
		t.Fatalf("expected reason 'grace-period', got %q", vc.Reason)
	}
}

func testPublicFeedSelfJudgingViewerContext(t *testing.T) {
	server, _ := newPublicReadTestServer(t)
	performJump(t, server, "alice-token")

	// Alice requests the feed and sees her own jump
	rec := doJSON(server, http.MethodGet, "/v1/feed", "alice-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res publicFeedBody
	decodeResponse(t, rec, &res)
	if len(res.Jumps) == 0 {
		t.Fatal("expected at least 1 jump in feed")
	}

	vc := res.Jumps[0].ViewerContext
	if vc == nil {
		t.Fatal("expected viewerContext on feed card for authenticated viewer")
	}
	if vc.CanJudge {
		t.Fatal("expected canJudge=false for self-judging jump on feed")
	}
	if vc.Reason != "self-judging" {
		t.Fatalf("expected reason 'self-judging', got %q", vc.Reason)
	}
}

func testPublicFeedOtherPlayerCanJudge(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performJump(t, server, "alice-token")

	// Advance past the Author Grace Period
	store.Store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	// Bob requests the feed — should be able to judge
	rec := doJSON(server, http.MethodGet, "/v1/feed", "bob-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res publicFeedBody
	decodeResponse(t, rec, &res)
	if len(res.Jumps) == 0 {
		t.Fatal("expected at least 1 jump in feed")
	}

	vc := res.Jumps[0].ViewerContext
	if vc == nil {
		t.Fatal("expected viewerContext on feed card for authenticated viewer")
	}
	if !vc.CanJudge {
		t.Fatal("expected canJudge=true for other player's jump after grace period")
	}
}

type failingPublicReadStore struct {
	*httpapi.PostgresStore
}

func (s *failingPublicReadStore) FeedJumps(_ context.Context, _ *time.Time, _ string, _ int) ([]httpapi.JumpCard, error) {
	return nil, errors.New("database unavailable")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// publicReadTestStore holds a pre-setup PostgresStore that tests can
// use to create Jumps and set the clock.
type publicReadTestStore struct {
	Store *httpapi.PostgresStore
	DB    httpapi.Persistence
}

type publicFeedBody struct {
	Jumps      []publicFeedJumpBody `json:"jumps"`
	NextCursor *string              `json:"nextCursor"`
}

type publicFeedJumpBody struct {
	ID             string  `json:"id"`
	PerformerName  string  `json:"performerName"`
	Source         string  `json:"source"`
	Destination    string  `json:"destination"`
	Food           string  `json:"food"`
	Caption        string  `json:"caption"`
	MediaObjectKey string  `json:"mediaObjectKey"`
	RunningAverage float64 `json:"runningAverage"`
	JudgmentCount  int     `json:"judgmentCount"`
	ViewerContext  *struct {
		CanJudge  bool   `json:"canJudge"`
		Reason    string `json:"reason,omitempty"`
		HasJudged bool   `json:"hasJudged"`
	} `json:"viewerContext"`
}

// newPublicReadTestServer creates a server + store with a clean database.
// Tests create Jumps directly via POST /v1/jumps.
func newPublicReadTestServer(t *testing.T) (http.Handler, *publicReadTestStore) {
	t.Helper()
	store := newCleanPostgresTestStore(t)
	server := newTestServerWithPersistence(store)
	return server, &publicReadTestStore{
		Store: store,
		DB:    store,
	}
}

// TestPublicFeedAlreadyJudged asserts that the feed shows already-judged state
// for authenticated viewers who have judged a jump.
func testPublicFeedAlreadyJudged(t *testing.T) {
	server, store := newPublicReadTestServer(t)
	performed := performJump(t, server, "alice-token")

	// Advance past the Author Grace Period before judging
	store.Store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	// Bob judges Alice's jump
	submitJudgment(t, server, "bob-token", performed.ID, 4, 3, 2, 1, http.StatusCreated)

	// Bob requests the feed — should see already-judged
	rec := doJSON(server, http.MethodGet, "/v1/feed", "bob-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res publicFeedBody
	decodeResponse(t, rec, &res)
	if len(res.Jumps) == 0 {
		t.Fatal("expected at least 1 jump in feed")
	}

	vc := res.Jumps[0].ViewerContext
	if vc == nil {
		t.Fatal("expected viewerContext on feed card for authenticated viewer")
	}
	if vc.CanJudge {
		t.Fatal("expected canJudge=false for already-judged jump on feed")
	}
	if !vc.HasJudged {
		t.Fatal("expected hasJudged=true for already-judged jump on feed")
	}
	if vc.Reason != "already-judged" {
		t.Fatalf("expected reason 'already-judged', got %q", vc.Reason)
	}
}

func TestPublicRead(t *testing.T) {
	t.Run("public feed", func(t *testing.T) {
		t.Run("returns jumps", testPublicFeedReturnsJumps)
		t.Run("returns multiple jumps in order", testPublicFeedReturnsMultipleJumpsInOrder)
		t.Run("cursor pagination", testPublicFeedCursorPagination)
		t.Run("invalid cursor returns 400", testPublicFeedInvalidCursorReturns400)
		t.Run("invalid limit returns 400", testPublicFeedInvalidLimitReturns400)
		t.Run("internal error returns message envelope", testPublicFeedInternalErrorReturnsMessageEnvelope)
		t.Run("grace period viewer context", testPublicFeedGracePeriodViewerContext)
		t.Run("self judging viewer context", testPublicFeedSelfJudgingViewerContext)
		t.Run("other player can judge", testPublicFeedOtherPlayerCanJudge)
		t.Run("already judged", testPublicFeedAlreadyJudged)
	})

	t.Run("public jump detail", func(t *testing.T) {
		t.Run("returns jump", testPublicJumpDetailReturnsJump)
		t.Run("unknown id returns 404", testPublicJumpDetailUnknownIDReturns404)
		t.Run("self judging viewer context", testPublicJumpDetailSelfJudgingViewerContext)
		t.Run("other player can judge", testPublicJumpDetailOtherPlayerCanJudge)
		t.Run("already judged viewer context", testPublicDetailAlreadyJudgedViewerContext)
		t.Run("grace period viewer context", testPublicJumpDetailGracePeriodViewerContext)
	})
}
