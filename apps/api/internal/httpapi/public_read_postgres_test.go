package httpapi_test

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresPublicFeedSurvivesRestartOrdersNewestFirstAndExcludesRemovedJumps(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Public Feed Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before public feed setup, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close Postgres database: %v", err)
		}
	})

	first := performCustomJump(t, server, "alice-token", group.Group.ID, "Crunchwrap", "First visible jump")
	if _, err := db.ExecContext(context.Background(), `UPDATE jumps SET grace_period_expires_at = now() - interval '1 minute' WHERE id = $1`, first.Jump.ID); err != nil {
		t.Fatalf("expire jump grace period: %v", err)
	}
	submitJudgment(t, server, "bob-token", first.Jump.ID, 4, 3, 2, 1, http.StatusCreated)

	time.Sleep(5 * time.Millisecond)
	removed := performCustomJump(t, server, "alice-token", group.Group.ID, "Burrito", "This jump should be removed")
	dispute := raiseDispute(t, server, "bob-token", removed.Jump.ID, "House Rules", "Remove this jump from the public feed")
	resolutionRec := doJSON(server, http.MethodPost, "/v1/disputes/"+dispute.ID+"/resolution", "alice-token", map[string]string{
		"resolution":       "Removed Jump",
		"resolutionReason": "Postgres public feed test removal",
	})
	if resolutionRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on dispute resolution, got %d: %s", resolutionRec.Code, resolutionRec.Body.String())
	}

	time.Sleep(5 * time.Millisecond)
	latest := performCustomJump(t, server, "alice-token", group.Group.ID, "Quesadilla", "Latest visible jump")

	restartedStore := newPostgresTestStore(t, databaseURL)
	restartedServer := newGroupsTestServerWithStore(restartedStore)

	rec := doJSON(restartedServer, http.MethodGet, "/v1/feed", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res publicFeedBody
	decodeResponse(t, rec, &res)
	latestIndex := -1
	firstIndex := -1
	for i, card := range res.Jumps {
		if card.ID == latest.Jump.ID {
			latestIndex = i
		}
		if card.ID == first.Jump.ID {
			firstIndex = i
			if card.Caption != first.Evidence.Caption {
				t.Fatalf("expected first visible caption %q, got %q", first.Evidence.Caption, card.Caption)
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
		t.Fatalf("expected latest visible jump %q in public feed", latest.Jump.ID)
	}
	if firstIndex == -1 {
		t.Fatalf("expected first visible jump %q in public feed", first.Jump.ID)
	}
	if latestIndex >= firstIndex {
		t.Fatalf("expected latest visible jump %q to appear before first visible jump %q, got indexes %d and %d", latest.Jump.ID, first.Jump.ID, latestIndex, firstIndex)
	}
	for _, card := range res.Jumps {
		if card.ID == removed.Jump.ID {
			t.Fatalf("expected removed jump %q to be excluded from Postgres feed", removed.Jump.ID)
		}
	}
}

func performCustomJump(t *testing.T, server http.Handler, token string, groupID string, food string, caption string) evidenceSubmissionBody {
	t.Helper()
	idea := createIdea(t, server, token, groupID, "Taco Bell", "Olive Garden parking lot", food)
	planned := createPlannedJump(t, server, token, idea.ID, false)
	authorization := authorizeEvidenceUpload(t, server, token, planned.ID, "image/jpeg")
	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence", token, map[string]string{
		"uploadAuthorizationId": authorization.ID,
		"caption":               caption,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var submission evidenceSubmissionBody
	decodeResponse(t, rec, &submission)
	normalizeEvidenceSubmissionBody(&submission)
	return submission
}
