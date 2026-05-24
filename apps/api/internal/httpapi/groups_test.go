package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

	inviteRec := doJSON(server, http.MethodPost, "/v1/groups/"+aliceGroup.Group.ID+"/invites", "alice-token", nil)
	if inviteRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", inviteRec.Code, inviteRec.Body.String())
	}
	var invite struct {
		ID      string `json:"id"`
		GroupID string `json:"groupId"`
		Token   string `json:"token"`
	}
	decodeResponse(t, inviteRec, &invite)
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
