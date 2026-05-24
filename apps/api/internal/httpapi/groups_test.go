package httpapi_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func TestCreateGroupMakesSignedInPlayerGroupAdminAndReturnsGroupHome(t *testing.T) {
	server := newGroupsTestServer()

	createRec := doJSON(server, http.MethodPost, "/v1/groups", "alice-token", map[string]string{"name": "Breakfast Crew"})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	var created struct {
		Group struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"group"`
		Membership struct {
			PlayerID string `json:"playerId"`
			Role     string `json:"role"`
		} `json:"membership"`
		ActiveSeason any   `json:"activeSeason"`
		RecentStunts []any `json:"recentStunts"`
		Standings    []any `json:"standings"`
	}
	decodeResponse(t, createRec, &created)

	if created.Group.ID == "" {
		t.Fatalf("expected created Group to have an id")
	}
	if created.Group.Name != "Breakfast Crew" {
		t.Fatalf("expected Group name from request, got %q", created.Group.Name)
	}
	if created.Membership.PlayerID == "" {
		t.Fatalf("expected Group Membership to identify the Player")
	}
	if created.Membership.Role != "Group Admin" {
		t.Fatalf("expected creator to become Group Admin, got %q", created.Membership.Role)
	}
	if created.ActiveSeason != nil {
		t.Fatalf("expected no Active Season, got %#v", created.ActiveSeason)
	}
	if len(created.RecentStunts) != 0 {
		t.Fatalf("expected no recent Stunts, got %#v", created.RecentStunts)
	}
	if len(created.Standings) != 0 {
		t.Fatalf("expected empty Standings, got %#v", created.Standings)
	}

	homeRec := doJSON(server, http.MethodGet, "/v1/groups/"+created.Group.ID+"/home", "alice-token", nil)
	if homeRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", homeRec.Code, homeRec.Body.String())
	}
	var home struct {
		Group struct {
			ID string `json:"id"`
		} `json:"group"`
		Membership struct {
			Role string `json:"role"`
		} `json:"membership"`
	}
	decodeResponse(t, homeRec, &home)
	if home.Group.ID != created.Group.ID || home.Membership.Role != "Group Admin" {
		t.Fatalf("expected backend Group home for created Group, got %#v", home)
	}
}

func TestSignedInPlayerCanListMultipleGroupsAndSwitchGroupHome(t *testing.T) {
	server := newGroupsTestServer()

	breakfast := createGroup(t, server, "alice-token", "Breakfast Crew")
	dinner := createGroup(t, server, "alice-token", "Dinner Weirdos")

	listRec := doJSON(server, http.MethodGet, "/v1/groups", "alice-token", nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", listRec.Code, listRec.Body.String())
	}
	var list struct {
		Memberships []struct {
			Group struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"group"`
			Membership struct {
				Role string `json:"role"`
			} `json:"membership"`
		} `json:"memberships"`
	}
	decodeResponse(t, listRec, &list)
	if len(list.Memberships) != 2 {
		t.Fatalf("expected two Group Memberships, got %#v", list.Memberships)
	}

	breakfastHome := getGroupHome(t, server, "alice-token", breakfast.Group.ID)
	dinnerHome := getGroupHome(t, server, "alice-token", dinner.Group.ID)
	if breakfastHome.Group.Name != "Breakfast Crew" {
		t.Fatalf("expected switched Group home for Breakfast Crew, got %#v", breakfastHome.Group)
	}
	if dinnerHome.Group.Name != "Dinner Weirdos" {
		t.Fatalf("expected switched Group home for Dinner Weirdos, got %#v", dinnerHome.Group)
	}
}

func TestGroupHomeRejectsSignedInNonMember(t *testing.T) {
	server := newGroupsTestServer()
	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodGet, "/v1/groups/"+aliceGroup.Group.ID+"/home", "bob-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGroupMemberCreatesInviteAndSignedInPlayerAcceptsWithoutReplacingExistingPlayHistory(t *testing.T) {
	server := newGroupsTestServer()
	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
	bobExistingGroup := createGroup(t, server, "bob-token", "Dinner Weirdos")

	invite := createInvite(t, server, "alice-token", aliceGroup.Group.ID)
	if invite.ID == "" || invite.GroupID != aliceGroup.Group.ID || invite.Token == "" {
		t.Fatalf("expected Invite for Alice's Group, got %#v", invite)
	}

	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	var accepted groupHomeBody
	decodeResponse(t, acceptRec, &accepted)
	if accepted.Group.ID != aliceGroup.Group.ID || accepted.Membership.Role != "Player" {
		t.Fatalf("expected Bob to join Alice's Group as Player, got %#v", accepted)
	}

	joinedHome := getGroupHome(t, server, "bob-token", aliceGroup.Group.ID)
	if joinedHome.Group.Name != "Breakfast Crew" || joinedHome.Membership.Role != "Player" {
		t.Fatalf("expected Bob to open invited Group home, got %#v", joinedHome)
	}
	stillOwnHome := getGroupHome(t, server, "bob-token", bobExistingGroup.Group.ID)
	if stillOwnHome.Group.Name != "Dinner Weirdos" || stillOwnHome.Membership.Role != "Group Admin" {
		t.Fatalf("expected Bob's existing play history to remain, got %#v", stillOwnHome)
	}
}

func TestAcceptInviteRejectsAlreadyUsedInvite(t *testing.T) {
	server := newGroupsTestServer()
	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", aliceGroup.Group.ID)

	first := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first accept status 200, got %d: %s", first.Code, first.Body.String())
	}

	second := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "carol-token", nil)
	if second.Code != http.StatusConflict {
		t.Fatalf("expected already-used Invite status 409, got %d: %s", second.Code, second.Body.String())
	}
}

func TestAcceptInviteRejectsExistingGroupMemberWithoutUsingInvite(t *testing.T) {
	server := newGroupsTestServer()
	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", aliceGroup.Group.ID)

	aliceAccept := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "alice-token", nil)
	if aliceAccept.Code != http.StatusConflict {
		t.Fatalf("expected existing member Invite accept status 409, got %d: %s", aliceAccept.Code, aliceAccept.Body.String())
	}

	bobAccept := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if bobAccept.Code != http.StatusOK {
		t.Fatalf("expected Invite to remain usable for Bob, got %d: %s", bobAccept.Code, bobAccept.Body.String())
	}
}

func TestAcceptInviteRejectsExpiredInvite(t *testing.T) {
	store := httpapi.NewMemoryStoreWithClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	})
	server := newGroupsTestServerWithStore(store)
	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", aliceGroup.Group.ID)

	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	})
	rec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if rec.Code != http.StatusGone {
		t.Fatalf("expected expired Invite status 410, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAcceptInviteRejectsInvalidInviteToken(t *testing.T) {
	server := newGroupsTestServer()

	rec := doJSON(server, http.MethodPost, "/v1/invites/not-a-real-invite/accept", "bob-token", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected invalid Invite status 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateInviteRejectsSignedInNonMember(t *testing.T) {
	server := newGroupsTestServer()
	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodPost, "/v1/groups/"+aliceGroup.Group.ID+"/invites", "bob-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected non-member Invite creation status 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGroupMemberCanStartSeasonAndSeeActiveSeasonOnGroupHome(t *testing.T) {
	server := newGroupsTestServer()
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	startRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", nil)
	if startRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", startRec.Code, startRec.Body.String())
	}

	var started groupHomeBody
	decodeResponse(t, startRec, &started)
	if started.ActiveSeason == nil {
		t.Fatalf("expected Active Season on start response")
	}
	if started.ActiveSeason.Status != "Active" {
		t.Fatalf("expected Active Season status, got %q", started.ActiveSeason.Status)
	}
	if started.ActiveSeason.CommissionerPlayerID != group.Membership.PlayerID {
		t.Fatalf("expected starting Player to be Season Commissioner, got %#v", started.ActiveSeason)
	}
	if len(started.Standings) != 0 {
		t.Fatalf("expected empty Standings, got %#v", started.Standings)
	}

	home := getGroupHome(t, server, "alice-token", group.Group.ID)
	if home.ActiveSeason == nil || home.ActiveSeason.ID != started.ActiveSeason.ID {
		t.Fatalf("expected Group home to show backend Active Season, got %#v", home.ActiveSeason)
	}
	if len(home.Standings) != 0 {
		t.Fatalf("expected empty Standings on Group home, got %#v", home.Standings)
	}
}

func TestGroupMemberCreatesIdeaThenSeasonLinkedPlannedStuntDuringActiveSeason(t *testing.T) {
	server := newGroupsTestServer()
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	season := startSeason(t, server, "alice-token", group.Group.ID)

	ideaRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/ideas", "alice-token", map[string]string{
		"source":      "Taco Bell",
		"destination": "Olive Garden parking lot",
		"food":        "Crunchwrap",
	})
	if ideaRec.Code != http.StatusCreated {
		t.Fatalf("expected Idea status 201, got %d: %s", ideaRec.Code, ideaRec.Body.String())
	}
	var idea stuntBody
	decodeResponse(t, ideaRec, &idea)
	if idea.ID == "" || idea.GroupID != group.Group.ID || idea.PlayerID != group.Membership.PlayerID {
		t.Fatalf("expected Group-scoped Idea for Alice, got %#v", idea)
	}
	if idea.Status != "Idea" || idea.SeasonID != nil || !idea.OffSeason {
		t.Fatalf("expected Off-Season Idea before planning, got %#v", idea)
	}

	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-stunt", "alice-token", nil)
	if planRec.Code != http.StatusCreated {
		t.Fatalf("expected Planned Stunt status 201, got %d: %s", planRec.Code, planRec.Body.String())
	}
	var planned stuntBody
	decodeResponse(t, planRec, &planned)
	if planned.ID != idea.ID || planned.Status != "Planned Stunt" {
		t.Fatalf("expected same Stunt to become a Planned Stunt, got %#v", planned)
	}
	if planned.GroupID != group.Group.ID {
		t.Fatalf("expected Planned Stunt to belong to exactly one Group %q, got %q", group.Group.ID, planned.GroupID)
	}
	if planned.SeasonID == nil || *planned.SeasonID != season.ActiveSeason.ID || planned.OffSeason {
		t.Fatalf("expected Planned Stunt to be Season-linked by default, got %#v", planned)
	}
	if planned.Source != "Taco Bell" || planned.Destination != "Olive Garden parking lot" || planned.Food != "Crunchwrap" {
		t.Fatalf("expected Source, Destination, and Food from Idea, got %#v", planned)
	}
}

func TestPlannedStuntCanBeExplicitlyOffSeasonDuringActiveSeason(t *testing.T) {
	server := newGroupsTestServer()
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	startSeason(t, server, "alice-token", group.Group.ID)
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Waffle House", "movie theater", "hash browns")

	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-stunt", "alice-token", map[string]bool{"offSeason": true})
	if planRec.Code != http.StatusCreated {
		t.Fatalf("expected Planned Stunt status 201, got %d: %s", planRec.Code, planRec.Body.String())
	}
	var planned stuntBody
	decodeResponse(t, planRec, &planned)
	if planned.SeasonID != nil || !planned.OffSeason {
		t.Fatalf("expected explicit Off-Season Stunt, got %#v", planned)
	}
}

func TestPlannedStuntIsOffSeasonWhenNoSeasonIsActive(t *testing.T) {
	server := newGroupsTestServer()
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Pizza Hut", "library", "personal pan pizza")

	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-stunt", "alice-token", nil)
	if planRec.Code != http.StatusCreated {
		t.Fatalf("expected Planned Stunt status 201, got %d: %s", planRec.Code, planRec.Body.String())
	}
	var planned stuntBody
	decodeResponse(t, planRec, &planned)
	if planned.SeasonID != nil || !planned.OffSeason {
		t.Fatalf("expected Off-Season Stunt without Active Season, got %#v", planned)
	}
}

func TestIdeaAndPlannedStuntRequireGroupMembership(t *testing.T) {
	server := newGroupsTestServer()
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Burger King", "bowling alley", "Whopper")

	createRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/ideas", "bob-token", map[string]string{
		"source":      "Taco Bell",
		"destination": "Olive Garden parking lot",
		"food":        "Crunchwrap",
	})
	if createRec.Code != http.StatusForbidden {
		t.Fatalf("expected non-member Idea creation status 403, got %d: %s", createRec.Code, createRec.Body.String())
	}

	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-stunt", "bob-token", nil)
	if planRec.Code != http.StatusForbidden {
		t.Fatalf("expected non-member Planned Stunt status 403, got %d: %s", planRec.Code, planRec.Body.String())
	}
}

func TestGroupCannotStartSecondSeasonWhileActiveSeasonExists(t *testing.T) {
	server := newGroupsTestServer()
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	firstRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", nil)
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	secondRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", nil)
	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
}

func TestPostgresGroupCreationSurvivesServerRestart(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	firstServer := newGroupsTestServerWithStore(store)
	created := createGroup(t, firstServer, "alice-token", "Breakfast Crew")

	restartedStore := newPostgresTestStore(t, databaseURL)
	restartedServer := newGroupsTestServerWithStore(restartedStore)
	home := getGroupHome(t, restartedServer, "alice-token", created.Group.ID)
	if home.Group.Name != "Breakfast Crew" || home.Membership.Role != "Group Admin" {
		t.Fatalf("expected durable Group Admin membership after restart, got %#v", home)
	}
}

func TestPostgresSeasonStartSurvivesRestartAndRejectsSecondOpenSeason(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	firstServer := newGroupsTestServerWithStore(store)
	group := createGroup(t, firstServer, "alice-token", "Breakfast Crew")
	started := startSeason(t, firstServer, "alice-token", group.Group.ID)

	restartedStore := newPostgresTestStore(t, databaseURL)
	restartedServer := newGroupsTestServerWithStore(restartedStore)
	home := getGroupHome(t, restartedServer, "alice-token", group.Group.ID)
	if home.ActiveSeason == nil || home.ActiveSeason.ID != started.ActiveSeason.ID {
		t.Fatalf("expected durable Active Season after restart, got %#v", home.ActiveSeason)
	}
	if home.ActiveSeason.CommissionerPlayerID != group.Membership.PlayerID {
		t.Fatalf("expected durable Season Commissioner, got %#v", home.ActiveSeason)
	}

	secondRec := doJSON(restartedServer, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", nil)
	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", secondRec.Code, secondRec.Body.String())
	}
}

func TestPostgresConcurrentSeasonStartsReturnCreatedAndConflict(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}
	installSlowSeasonInsertTrigger(t, databaseURL)

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	start := make(chan struct{})
	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			rec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", nil)
			statuses <- rec.Code
		}()
	}
	close(start)
	wg.Wait()
	close(statuses)

	seen := map[int]int{}
	for status := range statuses {
		seen[status]++
	}
	if seen[http.StatusCreated] != 1 || seen[http.StatusConflict] != 1 {
		t.Fatalf("expected one 201 and one 409, got %#v", seen)
	}
}

func TestPostgresConcurrentInviteCreationReturnsDistinctInvites(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	created := createGroup(t, server, "alice-token", "Concurrent Invite Creators")

	const inviteCount = 4
	invites := make(chan inviteBody, inviteCount)
	errors := make(chan string, inviteCount)
	var wg sync.WaitGroup
	for i := 0; i < inviteCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := doJSON(server, http.MethodPost, "/v1/groups/"+created.Group.ID+"/invites", "alice-token", nil)
			if rec.Code != http.StatusCreated {
				errors <- rec.Body.String()
				return
			}
			var invite inviteBody
			if err := json.NewDecoder(rec.Body).Decode(&invite); err != nil {
				errors <- err.Error()
				return
			}
			invites <- invite
		}()
	}
	wg.Wait()
	close(invites)
	close(errors)
	for err := range errors {
		t.Fatalf("expected concurrent Invite creation to succeed, got %s", err)
	}

	tokens := []string{}
	for invite := range invites {
		tokens = append(tokens, invite.Token)
	}
	if len(tokens) != inviteCount {
		t.Fatalf("expected %d Invites, got %d", inviteCount, len(tokens))
	}
	sort.Strings(tokens)
	for i := 1; i < len(tokens); i++ {
		if tokens[i] == tokens[i-1] {
			t.Fatalf("expected distinct Invite tokens, got %v", tokens)
		}
	}
}

func TestPostgresConcurrentInviteAcceptanceOnlyLetsOnePlayerConsumeInvite(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	created := createGroup(t, server, "alice-token", "Concurrent Invite Acceptors")
	invite := createInvite(t, server, "alice-token", created.Group.ID)

	codes := make(chan int, 2)
	var wg sync.WaitGroup
	for _, token := range []string{"bob-token", "carol-token"} {
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			rec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", token, nil)
			codes <- rec.Code
		}(token)
	}
	wg.Wait()
	close(codes)

	seen := map[int]int{}
	for code := range codes {
		seen[code]++
	}
	if seen[http.StatusOK] != 1 || seen[http.StatusConflict] != 1 {
		t.Fatalf("expected one winner and one already-used Invite conflict, got %#v", seen)
	}
}

func newGroupsTestServer() http.Handler {
	return newGroupsTestServerWithStore(httpapi.NewMemoryStore())
}

func newGroupsTestServerWithStore(store httpapi.Store) http.Handler {
	return httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "supabase", Subject: "alice-auth", Email: "alice@example.com"},
			"bob-token":   {Provider: "supabase", Subject: "bob-auth", Email: "bob@example.com"},
			"carol-token": {Provider: "supabase", Subject: "carol-auth", Email: "carol@example.com"},
		},
		Store: store,
	})
}

func newPostgresTestStore(t *testing.T, databaseURL string) httpapi.Store {
	t.Helper()
	store, err := httpapi.NewPostgresStore(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("new Postgres store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close Postgres store: %v", err)
		}
	})
	return store
}

func installSlowSeasonInsertTrigger(t *testing.T, databaseURL string) {
	t.Helper()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close Postgres database: %v", err)
		}
	})
	if _, err := db.ExecContext(context.Background(), `
CREATE OR REPLACE FUNCTION supperjumpin_test_slow_season_insert()
RETURNS trigger AS $$
BEGIN
  PERFORM pg_sleep(0.2);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS supperjumpin_test_slow_season_insert ON seasons;
CREATE TRIGGER supperjumpin_test_slow_season_insert
BEFORE INSERT ON seasons
FOR EACH ROW EXECUTE FUNCTION supperjumpin_test_slow_season_insert();`); err != nil {
		t.Fatalf("install slow Season insert trigger: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `
DROP TRIGGER IF EXISTS supperjumpin_test_slow_season_insert ON seasons;
DROP FUNCTION IF EXISTS supperjumpin_test_slow_season_insert();`); err != nil {
			t.Fatalf("remove slow Season insert trigger: %v", err)
		}
	})
}

func doJSON(server http.Handler, method string, path string, token string, body any) *httptest.ResponseRecorder {
	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			panic(err)
		}
	}
	req := httptest.NewRequest(method, path, &requestBody)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

type groupHomeBody struct {
	Group struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"group"`
	Membership struct {
		PlayerID string `json:"playerId"`
		Role     string `json:"role"`
	} `json:"membership"`
	ActiveSeason *struct {
		ID                   string `json:"id"`
		Status               string `json:"status"`
		CommissionerPlayerID string `json:"commissionerPlayerId"`
	} `json:"activeSeason"`
	Standings []any `json:"standings"`
}

func createGroup(t *testing.T, server http.Handler, token string, name string) groupHomeBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/groups", token, map[string]string{"name": name})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body groupHomeBody
	decodeResponse(t, rec, &body)
	return body
}

func startSeason(t *testing.T, server http.Handler, token string, groupID string) groupHomeBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/groups/"+groupID+"/seasons", token, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body groupHomeBody
	decodeResponse(t, rec, &body)
	return body
}

func getGroupHome(t *testing.T, server http.Handler, token string, groupID string) groupHomeBody {
	t.Helper()
	rec := doJSON(server, http.MethodGet, "/v1/groups/"+groupID+"/home", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body groupHomeBody
	decodeResponse(t, rec, &body)
	return body
}

type inviteBody struct {
	ID      string `json:"id"`
	GroupID string `json:"groupId"`
	Token   string `json:"token"`
}

type stuntBody struct {
	ID          string  `json:"id"`
	GroupID     string  `json:"groupId"`
	PlayerID    string  `json:"playerId"`
	SeasonID    *string `json:"seasonId"`
	Status      string  `json:"status"`
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	Food        string  `json:"food"`
	OffSeason   bool    `json:"offSeason"`
}

func createInvite(t *testing.T, server http.Handler, token string, groupID string) inviteBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/groups/"+groupID+"/invites", token, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body inviteBody
	decodeResponse(t, rec, &body)
	return body
}

func createIdea(t *testing.T, server http.Handler, token string, groupID string, source string, destination string, food string) stuntBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/groups/"+groupID+"/ideas", token, map[string]string{
		"source":      source,
		"destination": destination,
		"food":        food,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body stuntBody
	decodeResponse(t, rec, &body)
	return body
}
