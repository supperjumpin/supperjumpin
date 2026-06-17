package httpapi_test

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func testPostgresPublicFeedSurvivesRestartOrdersNewestFirst(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close Postgres database: %v", err)
		}
	})

	first := performCustomJump(t, server, "alice-token", "Crunchwrap", "First visible jump")
	if _, err := db.ExecContext(context.Background(), `UPDATE jumps SET grace_period_expires_at = now() - interval '1 minute' WHERE id = $1`, first.ID); err != nil {
		t.Fatalf("expire jump grace period: %v", err)
	}
	submitJudgment(t, server, "bob-token", first.ID, 4, 3, 2, 1, http.StatusCreated)

	time.Sleep(5 * time.Millisecond)
	latest := performCustomJump(t, server, "alice-token", "Quesadilla", "Latest visible jump")

	restartedStore := newPostgresTestStore(t, databaseURL)
	restartedServer := newTestServerWithStore(restartedStore)

	rec := doJSON(restartedServer, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res publicFeedBody
	decodeResponse(t, rec, &res)
	latestIndex := -1
	firstIndex := -1
	for i, card := range res.Jumps {
		if card.ID == latest.ID {
			latestIndex = i
		}
		if card.ID == first.ID {
			firstIndex = i
			if card.Caption != "First visible jump" {
				t.Fatalf("expected first visible caption %q, got %q", "First visible jump", card.Caption)
			}
			if card.ViewerContext != nil {
				t.Fatal("expected unauthenticated Postgres feed card to omit viewerContext")
			}
			if card.JudgmentCount != 1 {
				t.Fatalf("expected first visible jump judgmentCount 1, got %d", card.JudgmentCount)
			}
			if card.RunningAverage != 2.5 {
				t.Fatalf("expected first visible jump runningAverage 2.5, got %v", card.RunningAverage)
			}
		}
	}
	if latestIndex == -1 {
		t.Fatalf("expected latest visible jump %q in public feed", latest.ID)
	}
	if firstIndex == -1 {
		t.Fatalf("expected first visible jump %q in public feed", first.ID)
	}
	if latestIndex >= firstIndex {
		t.Fatalf("expected latest visible jump %q to appear before first visible jump %q, got indexes %d and %d", latest.ID, first.ID, latestIndex, firstIndex)
	}
}

func testPostgresPublicJumpDetailCoversVisibleJumpAndMissingJump(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)
	performed := performJump(t, server, "alice-token")

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close postgres database: %v", err)
		}
	})
	if _, err := db.ExecContext(context.Background(), `UPDATE jumps SET grace_period_expires_at = now() - interval '1 minute' WHERE id = $1`, performed.ID); err != nil {
		t.Fatalf("expire jump grace period: %v", err)
	}

	detailRec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "", nil)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", detailRec.Code, detailRec.Body.String())
	}

	var detail struct {
		ID                   string    `json:"id"`
		PerformerName        string    `json:"performerName"`
		PerformerID          string    `json:"performerId"`
		Source               string    `json:"source"`
		Destination          string    `json:"destination"`
		Food                 string    `json:"food"`
		Caption              string    `json:"caption"`
		MediaObjectKey       string    `json:"mediaObjectKey"`
		Status               string    `json:"status"`
		GracePeriodExpiresAt time.Time `json:"gracePeriodExpiresAt"`
		RunningAverage       float64   `json:"runningAverage"`
		JudgmentCount        int       `json:"judgmentCount"`
		CreatedAt            time.Time `json:"createdAt"`
		ViewerContext        *struct {
			CanJudge bool `json:"canJudge"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, detailRec, &detail)
	if detail.ID != performed.ID {
		t.Fatalf("expected jump ID %q, got %q", performed.ID, detail.ID)
	}
	if detail.PerformerName == "" {
		t.Fatal("expected performerName to be populated")
	}
	if detail.PerformerID == "" {
		t.Fatal("expected performerId to be populated")
	}
	if detail.Source != "Taco Bell" || detail.Destination != "Olive Garden parking lot" || detail.Food != "Crunchwrap" {
		t.Fatalf("expected source/destination/food in detail, got %+v", detail)
	}
	if detail.Caption == "" {
		t.Fatal("expected caption to be populated")
	}
	if detail.MediaObjectKey == "" {
		t.Fatal("expected mediaObjectKey to be populated")
	}
	if detail.Status != "Performed Jump" {
		t.Fatalf("expected status 'Performed Jump', got %q", detail.Status)
	}
	if detail.GracePeriodExpiresAt.IsZero() {
		t.Fatal("expected gracePeriodExpiresAt to be populated")
	}
	if detail.CreatedAt.IsZero() {
		t.Fatal("expected createdAt to be populated")
	}
	if detail.RunningAverage != 0 {
		t.Fatalf("expected runningAverage 0 before judgments, got %v", detail.RunningAverage)
	}
	if detail.JudgmentCount != 0 {
		t.Fatalf("expected judgmentCount 0 before judgments, got %d", detail.JudgmentCount)
	}
	if detail.ViewerContext == nil || !detail.ViewerContext.CanJudge {
		t.Fatal("expected anonymous viewer to see judge CTA")
	}

	unknownRec := doJSON(server, http.MethodGet, "/v1/jumps/not-found", "", nil)
	if unknownRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown jump, got %d: %s", unknownRec.Code, unknownRec.Body.String())
	}
	var missing struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	decodeResponse(t, unknownRec, &missing)
	if missing.Error != "not_found" {
		t.Fatalf("expected not_found error code, got %q", missing.Error)
	}
	if missing.Message != "Jump not found. It may have been removed." {
		t.Fatalf("expected jump not found message, got %q", missing.Message)
	}

	submitJudgment(t, server, "bob-token", performed.ID, 4, 3, 2, 1, http.StatusCreated)
	judgedRec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "bob-token", nil)
	if judgedRec.Code != http.StatusOK {
		t.Fatalf("expected 200 after judgment, got %d: %s", judgedRec.Code, judgedRec.Body.String())
	}

	var judged struct {
		RunningAverage float64 `json:"runningAverage"`
		JudgmentCount  int     `json:"judgmentCount"`
		ViewerContext  *struct {
			CanJudge  bool   `json:"canJudge"`
			Reason    string `json:"reason,omitempty"`
			HasJudged bool   `json:"hasJudged"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, judgedRec, &judged)
	if judged.RunningAverage != 2.5 {
		t.Fatalf("expected runningAverage 2.5 after one judgment, got %v", judged.RunningAverage)
	}
	if judged.JudgmentCount != 1 {
		t.Fatalf("expected judgmentCount 1 after one judgment, got %d", judged.JudgmentCount)
	}
	if judged.ViewerContext == nil {
		t.Fatal("expected viewerContext for signed-in viewer")
	}
	if judged.ViewerContext.CanJudge {
		t.Fatal("expected already-judged viewer to not be able to judge")
	}
	if judged.ViewerContext.Reason != "already-judged" {
		t.Fatalf("expected reason 'already-judged', got %q", judged.ViewerContext.Reason)
	}
	if !judged.ViewerContext.HasJudged {
		t.Fatal("expected hasJudged=true after judgment")
	}
}



func testPostgresPublicJumpDetailGracePeriodViewerContext(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)
	performed := performJump(t, server, "alice-token")

	// Grace period is naturally active after performJump; Bob views detail
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "bob-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var detail struct {
		ViewerContext *struct {
			CanJudge          bool    `json:"canJudge"`
			Reason            string  `json:"reason,omitempty"`
			GracePeriodEndsAt *string `json:"gracePeriodEndsAt,omitempty"`
		} `json:"viewerContext"`
	}
	decodeResponse(t, rec, &detail)
	if detail.ViewerContext == nil {
		t.Fatal("expected viewerContext")
	}
	if detail.ViewerContext.CanJudge {
		t.Fatal("expected canJudge=false during grace period")
	}
	if detail.ViewerContext.Reason != "grace-period" {
		t.Fatalf("expected reason 'grace-period', got %q", detail.ViewerContext.Reason)
	}
	if detail.ViewerContext.GracePeriodEndsAt == nil || *detail.ViewerContext.GracePeriodEndsAt == "" {
		t.Fatal("expected gracePeriodEndsAt to be populated during grace period")
	}
}

func testPostgresPublicJumpDetailSelfJudgingViewerContext(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)
	performed := performJump(t, server, "alice-token")

	// Alice views her own jump detail — self-judging should take precedence over grace period
	rec := doJSON(server, http.MethodGet, "/v1/jumps/"+performed.ID, "alice-token", nil)
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
	if detail.ViewerContext.CanJudge {
		t.Fatal("expected canJudge=false for self-judging")
	}
	if detail.ViewerContext.Reason != "self-judging" {
		t.Fatalf("expected reason 'self-judging', got %q", detail.ViewerContext.Reason)
	}
}

func testPostgresPublicFeedGracePeriodViewerContext(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)
	performJump(t, server, "alice-token")

	// Grace period is naturally active; Bob requests the feed
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

func testPostgresPublicFeedSelfJudgingViewerContext(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)
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

func performCustomJump(t *testing.T, server http.Handler, token string, food string, caption string) jumpBody {
	t.Helper()
	mediaObjectKey := "evidence_object_" + strconv.Itoa(int(time.Now().UnixNano()))
	rec := doJSON(server, http.MethodPost, "/v1/jumps", token, map[string]string{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           food,
		"caption":        caption,
		"mediaObjectKey": mediaObjectKey,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body jumpBody
	decodeResponse(t, rec, &body)
	return body
}

func testPostgresPublicFeedEmptyReturnsEmptyArrayAndNullCursor(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Postgres database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close Postgres database: %v", err)
		}
	}()
	cleanTestDatabase(t, db)

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)

	rec := doJSON(server, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var res publicFeedBody
	decodeResponse(t, rec, &res)
	if len(res.Jumps) != 0 {
		t.Fatalf("expected empty feed, got %d jumps", len(res.Jumps))
	}
	if res.NextCursor != nil {
		t.Fatalf("expected nil cursor for empty feed, got %q", *res.NextCursor)
	}
}

func testPostgresPublicFeedCursorPaginationMultiPage(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Postgres database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close Postgres database: %v", err)
		}
	}()
	cleanTestDatabase(t, db)

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)

	var created []string
	for i := 0; i < 25; i++ {
		time.Sleep(5 * time.Millisecond)
		jump := performCustomJump(t, server, "alice-token", "Food"+strconv.Itoa(i), "Jump "+strconv.Itoa(i))
		created = append(created, jump.ID)
	}

	rec := doJSON(server, http.MethodGet, "/v1/feed?limit=10", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var page1 publicFeedBody
	decodeResponse(t, rec, &page1)
	if len(page1.Jumps) != 10 {
		t.Fatalf("expected 10 on page 1, got %d", len(page1.Jumps))
	}
	if page1.NextCursor == nil {
		t.Fatal("expected nextCursor on page 1")
	}
	if page1.Jumps[0].ID != created[24] {
		t.Fatalf("expected newest jump first")
	}

	rec2 := doJSON(server, http.MethodGet, "/v1/feed?cursor="+*page1.NextCursor+"&limit=10", "", nil)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for page 2, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var page2 publicFeedBody
	decodeResponse(t, rec2, &page2)
	if len(page2.Jumps) != 10 {
		t.Fatalf("expected 10 on page 2, got %d", len(page2.Jumps))
	}
	if page2.NextCursor == nil {
		t.Fatal("expected nextCursor on page 2")
	}

	rec3 := doJSON(server, http.MethodGet, "/v1/feed?cursor="+*page2.NextCursor+"&limit=10", "", nil)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 for page 3, got %d: %s", rec3.Code, rec3.Body.String())
	}
	var page3 publicFeedBody
	decodeResponse(t, rec3, &page3)
	if len(page3.Jumps) != 5 {
		t.Fatalf("expected 5 on page 3, got %d", len(page3.Jumps))
	}
	if page3.NextCursor != nil {
		t.Fatal("expected nil cursor on last page")
	}

	seen := map[string]bool{}
	for _, p := range []publicFeedBody{page1, page2, page3} {
		for _, j := range p.Jumps {
			if seen[j.ID] {
				t.Fatalf("duplicate jump %q across pages", j.ID)
			}
			seen[j.ID] = true
		}
	}
	if len(seen) != 25 {
		t.Fatalf("expected 25 unique jumps, got %d", len(seen))
	}
}

func testPostgresPublicFeedSameTimestampTiebrokenByID(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Postgres database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close Postgres database: %v", err)
		}
	}()
	cleanTestDatabase(t, db)

	store := newPostgresTestStore(t, databaseURL)
	server := newTestServerWithStore(store)

	var created []string
	for i := 0; i < 5; i++ {
		time.Sleep(5 * time.Millisecond)
		jump := performCustomJump(t, server, "alice-token", "Food"+strconv.Itoa(i), "Jump "+strconv.Itoa(i))
		created = append(created, jump.ID)
	}

	// Force all created_at to the same value
	sameTime := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	for _, id := range created {
		if _, err := db.ExecContext(context.Background(), `UPDATE jumps SET created_at = $1 WHERE id = $2`, sameTime, id); err != nil {
			t.Fatalf("set same created_at: %v", err)
		}
	}

	rec1 := doJSON(server, http.MethodGet, "/v1/feed?limit=3", "", nil)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec1.Code, rec1.Body.String())
	}
	var page1 publicFeedBody
	decodeResponse(t, rec1, &page1)
	if len(page1.Jumps) != 3 {
		t.Fatalf("expected 3 on page 1, got %d", len(page1.Jumps))
	}
	if page1.NextCursor == nil {
		t.Fatal("expected nextCursor on page 1")
	}

	rec1b := doJSON(server, http.MethodGet, "/v1/feed?limit=3", "", nil)
	if rec1b.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec1b.Code, rec1b.Body.String())
	}
	var page1b publicFeedBody
	decodeResponse(t, rec1b, &page1b)
	for i := 0; i < 3; i++ {
		if page1.Jumps[i].ID != page1b.Jumps[i].ID {
			t.Fatalf("deterministic order violated at index %d: %q vs %q", i, page1.Jumps[i].ID, page1b.Jumps[i].ID)
		}
	}

	rec2 := doJSON(server, http.MethodGet, "/v1/feed?cursor="+*page1.NextCursor+"&limit=3", "", nil)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for page 2, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var page2 publicFeedBody
	decodeResponse(t, rec2, &page2)
	if len(page2.Jumps) != 2 {
		t.Fatalf("expected 2 on page 2, got %d", len(page2.Jumps))
	}
	if page2.NextCursor != nil {
		t.Fatal("expected nil cursor on last page")
	}

	seen := map[string]bool{}
	for _, j := range page1.Jumps {
		seen[j.ID] = true
	}
	for _, j := range page2.Jumps {
		if seen[j.ID] {
			t.Fatalf("duplicate jump %q across pages", j.ID)
		}
		seen[j.ID] = true
	}
	if len(seen) != 5 {
		t.Fatalf("expected 5 unique jumps, got %d", len(seen))
	}
}

func TestPostgresPublicRead(t *testing.T) {
	t.Run("public feed", func(t *testing.T) {
		t.Run("survives restart orders newest first", testPostgresPublicFeedSurvivesRestartOrdersNewestFirst)
		t.Run("grace period viewer context", testPostgresPublicFeedGracePeriodViewerContext)
		t.Run("self judging viewer context", testPostgresPublicFeedSelfJudgingViewerContext)
		t.Run("empty returns empty array and null cursor", testPostgresPublicFeedEmptyReturnsEmptyArrayAndNullCursor)
		t.Run("cursor pagination multi page", testPostgresPublicFeedCursorPaginationMultiPage)
		t.Run("same timestamp tiebroken by id", testPostgresPublicFeedSameTimestampTiebrokenByID)
	})

	t.Run("public jump detail", func(t *testing.T) {
		t.Run("covers visible jump and missing jump", testPostgresPublicJumpDetailCoversVisibleJumpAndMissingJump)
		t.Run("grace period viewer context", testPostgresPublicJumpDetailGracePeriodViewerContext)
		t.Run("self judging viewer context", testPostgresPublicJumpDetailSelfJudgingViewerContext)
	})
}
