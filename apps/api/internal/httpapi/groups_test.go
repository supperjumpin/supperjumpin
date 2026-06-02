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
	server := newGroupsTestServer(t)

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
		RecentJumps  []any `json:"recentJumps"`
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
	if len(created.RecentJumps) != 0 {
		t.Fatalf("expected no recent Jumps, got %#v", created.RecentJumps)
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
	server := newGroupsTestServer(t)

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
	server := newGroupsTestServer(t)
	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodGet, "/v1/groups/"+aliceGroup.Group.ID+"/home", "bob-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGroupMemberCreatesInviteAndSignedInPlayerAcceptsWithoutReplacingExistingPlayHistory(t *testing.T) {
	server := newGroupsTestServer(t)
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
	server := newGroupsTestServer(t)
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
	server := newGroupsTestServer(t)
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
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
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
	server := newGroupsTestServer(t)

	rec := doJSON(server, http.MethodPost, "/v1/invites/not-a-real-invite/accept", "bob-token", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected invalid Invite status 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAcceptInviteReturnsStandingsForFinalizedSeason(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	})
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	)
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 11, 0, 0, time.UTC)
	})

	submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 3, 2, http.StatusCreated)

	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	})

	carolAcceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+createInvite(t, server, "alice-token", group.Group.ID).Token+"/accept", "carol-token", nil)
	if carolAcceptRec.Code != http.StatusOK {
		t.Fatalf("expected Carol to join Group after judging, got %d: %s", carolAcceptRec.Code, carolAcceptRec.Body.String())
	}
	var accepted groupHomeBody
	decodeResponse(t, carolAcceptRec, &accepted)
	if len(accepted.Standings) != 1 {
		t.Fatalf("expected standings in Invite accept response, got %#v", accepted.Standings)
	}
	if accepted.Standings[0].Player.ID != group.Membership.PlayerID || accepted.Standings[0].SeasonScore != 12 || accepted.Standings[0].JudgedJumps != 1 {
		t.Fatalf("expected Alice standings in Invite accept response, got %#v", accepted.Standings[0])
	}
}

func TestCreateInviteRejectsSignedInNonMember(t *testing.T) {
	server := newGroupsTestServer(t)
	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodPost, "/v1/groups/"+aliceGroup.Group.ID+"/invites", "bob-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected non-member Invite creation status 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGroupMemberCanStartSeasonAndSeeActiveSeasonOnGroupHome(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	startRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", map[string]string{
		"submissionDeadline": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
		"judgingDeadline":    time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
	})
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

func TestSeasonCommissionerCanCloseSubmissionsAndMoveSeasonIntoJudgingGracePeriod(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	season := startSeason(t, server, "alice-token", group.Group.ID)
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Waffle House", "movie theater", "hash browns")
	planned := createPlannedJump(t, server, "alice-token", idea.ID, false)
	authorization := authorizeEvidenceUpload(t, server, "alice-token", planned.ID, "image/jpeg")

	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "alice-token", nil)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
	}
	var closed groupHomeBody
	decodeResponse(t, closeRec, &closed)
	if closed.ActiveSeason == nil || closed.ActiveSeason.ID != season.ActiveSeason.ID || closed.ActiveSeason.Status != "Judging Grace Period" {
		t.Fatalf("expected Season to move into Judging Grace Period, got %#v", closed.ActiveSeason)
	}

	submitRec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence", "alice-token", map[string]string{
		"uploadAuthorizationId": authorization.ID,
		"caption":               "Too late for competition.",
	})
	if submitRec.Code != http.StatusConflict {
		t.Fatalf("expected closed submission status 409, got %d: %s", submitRec.Code, submitRec.Body.String())
	}
}

func TestJudgingGracePeriodStillAllowsEligibleJudgmentsOnExistingPerformedJumps(t *testing.T) {
	server, store := newGroupsTestServerAndStore(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	season := startSeason(t, server, "alice-token", group.Group.ID)
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "alice-token", nil)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
	}

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	judgment := submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 3, 2, http.StatusCreated)
	if judgment.JumpID != performed.Jump.ID || judgment.PlayerID == performed.Jump.PlayerID {
		t.Fatalf("expected eligible Judge to score existing Performed Jump during Judging Grace Period, got %#v", judgment)
	}
}

func TestSeasonCommissionerCanFinalizeSeasonAndLockStandings(t *testing.T) {
	server, store := newGroupsTestServerAndStore(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	season := startSeason(t, server, "alice-token", group.Group.ID)
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 3, 2, http.StatusCreated)

	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "alice-token", nil)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
	}

	finalizeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/finalize", "alice-token", nil)
	if finalizeRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", finalizeRec.Code, finalizeRec.Body.String())
	}
	var finalized groupHomeBody
	decodeResponse(t, finalizeRec, &finalized)
	if finalized.ActiveSeason != nil {
		t.Fatalf("expected Finalized Season to no longer be open, got %#v", finalized.ActiveSeason)
	}
	if len(finalized.Standings) != 1 || finalized.Standings[0].SeasonScore != 12 || finalized.Standings[0].JudgedJumps != 1 {
		t.Fatalf("expected locked Standings with Alice on 12 from one Judged Jump, got %#v", finalized.Standings)
	}

	editRec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
		"commitment":    4,
		"transgression": 4,
		"creativity":    4,
		"presentation":  4,
	})
	if editRec.Code != http.StatusConflict {
		t.Fatalf("expected Finalized Season to lock remaining Judging Windows, got %d: %s", editRec.Code, editRec.Body.String())
	}
}

func TestGroupAdminEmergencySeasonOverridesAppearInSeasonHistory(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "bob-token", "Breakfast Crew")
	invite := createInvite(t, server, "bob-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "alice-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Alice to join Group before starting Season, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	season := startSeason(t, server, "alice-token", group.Group.ID)
	performed := performJump(t, server, "alice-token", group.Group.ID)

	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "bob-token", nil)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("expected Group Admin override close status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
	}

	finalizeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/finalize", "bob-token", nil)
	if finalizeRec.Code != http.StatusOK {
		t.Fatalf("expected Group Admin override finalize status 200, got %d: %s", finalizeRec.Code, finalizeRec.Body.String())
	}

	history := getSeasonHistory(t, server, "alice-token", season.ActiveSeason.ID)
	if len(history.Entries) != 2 {
		t.Fatalf("expected two visible override actions in Season history, got %#v", history.Entries)
	}
	if history.Entries[0].Action != "Submissions Closed" || history.Entries[0].ActorPlayerID != group.Membership.PlayerID || !history.Entries[0].Override || history.Entries[0].ActorRole != "Group Admin" {
		t.Fatalf("expected visible Group Admin close override entry, got %#v", history.Entries[0])
	}
	if history.Entries[1].Action != "Season Finalized" || history.Entries[1].ActorPlayerID != group.Membership.PlayerID || !history.Entries[1].Override || history.Entries[1].ActorRole != "Group Admin" {
		t.Fatalf("expected visible Group Admin finalize override entry, got %#v", history.Entries[1])
	}
	if history.Entries[1].ToStatus != "Finalized" {
		t.Fatalf("expected finalized status in Season history, got %#v", history.Entries[1])
	}

	editRec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
		"commitment":    1,
		"transgression": 1,
		"creativity":    1,
		"presentation":  1,
	})
	if editRec.Code != http.StatusConflict {
		t.Fatalf("expected Group Admin finalization override to cap remaining Judging Windows, got %d: %s", editRec.Code, editRec.Body.String())
	}
}

func TestGroupMemberCreatesIdeaThenSeasonLinkedPlannedJumpDuringActiveSeason(t *testing.T) {
	server := newGroupsTestServer(t)
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
	var idea jumpBody
	decodeResponse(t, ideaRec, &idea)
	if idea.ID == "" || idea.GroupID != group.Group.ID || idea.PlayerID != group.Membership.PlayerID {
		t.Fatalf("expected Group-scoped Idea for Alice, got %#v", idea)
	}
	if idea.Status != "Idea" || idea.SeasonID != nil || !idea.OffSeason {
		t.Fatalf("expected Off-Season Idea before planning, got %#v", idea)
	}

	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-jump", "alice-token", nil)
	if planRec.Code != http.StatusCreated {
		t.Fatalf("expected Planned Jump status 201, got %d: %s", planRec.Code, planRec.Body.String())
	}
	var planned jumpBody
	decodeResponse(t, planRec, &planned)
	if planned.ID != idea.ID || planned.Status != "Planned Jump" {
		t.Fatalf("expected same Jump to become a Planned Jump, got %#v", planned)
	}
	if planned.GroupID != group.Group.ID {
		t.Fatalf("expected Planned Jump to belong to exactly one Group %q, got %q", group.Group.ID, planned.GroupID)
	}
	if planned.SeasonID == nil || *planned.SeasonID != season.ActiveSeason.ID || planned.OffSeason {
		t.Fatalf("expected Planned Jump to be Season-linked by default, got %#v", planned)
	}
	if planned.Source != "Taco Bell" || planned.Destination != "Olive Garden parking lot" || planned.Food != "Crunchwrap" {
		t.Fatalf("expected Source, Destination, and Food from Idea, got %#v", planned)
	}
}

func TestPlannedJumpCanBeExplicitlyOffSeasonDuringActiveSeason(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	startSeason(t, server, "alice-token", group.Group.ID)
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Waffle House", "movie theater", "hash browns")

	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-jump", "alice-token", map[string]bool{"offSeason": true})
	if planRec.Code != http.StatusCreated {
		t.Fatalf("expected Planned Jump status 201, got %d: %s", planRec.Code, planRec.Body.String())
	}
	var planned jumpBody
	decodeResponse(t, planRec, &planned)
	if planned.SeasonID != nil || !planned.OffSeason {
		t.Fatalf("expected explicit Off-Season Jump, got %#v", planned)
	}
}

func TestPlannedJumpIsOffSeasonWhenNoSeasonIsActive(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Pizza Hut", "library", "personal pan pizza")

	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-jump", "alice-token", nil)
	if planRec.Code != http.StatusCreated {
		t.Fatalf("expected Planned Jump status 201, got %d: %s", planRec.Code, planRec.Body.String())
	}
	var planned jumpBody
	decodeResponse(t, planRec, &planned)
	if planned.SeasonID != nil || !planned.OffSeason {
		t.Fatalf("expected Off-Season Jump without Active Season, got %#v", planned)
	}
}

func TestPlannedJumpPerformerCanAuthorizeEvidenceUpload(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	createGroup(t, server, "bob-token", "Side Judges")
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Taco Bell", "Olive Garden parking lot", "Crunchwrap")
	planned := createPlannedJump(t, server, "alice-token", idea.ID, false)

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence-upload-authorizations", "alice-token", map[string]string{
		"contentType": "image/jpeg",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var authorization evidenceUploadAuthorizationBody
	decodeResponse(t, rec, &authorization)
	if authorization.ID == "" {
		t.Fatalf("expected authorization id, got %#v", authorization)
	}
	if authorization.JumpID != planned.ID {
		t.Fatalf("expected authorization for Planned Jump %q, got %#v", planned.ID, authorization)
	}
	if authorization.UploadMethod != "PUT" {
		t.Fatalf("expected direct upload method PUT, got %#v", authorization)
	}
	if authorization.UploadURL == "" || authorization.MediaObjectKey == "" {
		t.Fatalf("expected direct upload target fields, got %#v", authorization)
	}
	if authorization.UploadHeaders["Content-Type"] != "image/jpeg" {
		t.Fatalf("expected upload content type header, got %#v", authorization.UploadHeaders)
	}
}

func TestEvidenceUploadAuthorizationRejectsNonPerformer(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before authorization attempt, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Taco Bell", "Olive Garden parking lot", "Crunchwrap")
	planned := createPlannedJump(t, server, "alice-token", idea.ID, false)

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence-upload-authorizations", "bob-token", map[string]string{
		"contentType": "image/jpeg",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthorizedEvidenceSubmissionPerformsJumpAndOwnsMediaObjectKey(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Taco Bell", "Olive Garden parking lot", "Crunchwrap")
	planned := createPlannedJump(t, server, "alice-token", idea.ID, false)
	authorization := authorizeEvidenceUpload(t, server, "alice-token", planned.ID, "image/jpeg")

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence", "alice-token", map[string]string{
		"uploadAuthorizationId": authorization.ID,
		"caption":               "Crunchwrap successfully smuggled into the parking lot.",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var submission evidenceSubmissionBody
	decodeResponse(t, rec, &submission)
	if submission.Jump.Status != "Performed Jump" {
		t.Fatalf("expected Planned Jump to become Performed Jump, got %#v", submission.Jump)
	}
	if submission.Evidence.Caption != "Crunchwrap successfully smuggled into the parking lot." {
		t.Fatalf("expected caption to be stored, got %#v", submission.Evidence)
	}
	if submission.Evidence.MediaObjectKey != authorization.MediaObjectKey {
		t.Fatalf("expected backend-owned media object key from authorization, got %#v", submission.Evidence)
	}

	home := getGroupHome(t, server, "alice-token", group.Group.ID)
	if len(home.RecentJumps) != 1 {
		t.Fatalf("expected one recent Performed Jump on Group home, got %#v", home.RecentJumps)
	}
	recent := home.RecentJumps[0]
	if recent.Jump.ID != planned.ID || recent.Jump.Status != "Performed Jump" {
		t.Fatalf("expected recent Performed Jump to match the submitted Jump, got %#v", recent)
	}
	if recent.Performer.DisplayName != "alice" {
		t.Fatalf("expected performer display name on recent Jump, got %#v", recent.Performer)
	}
	if recent.Evidence.Caption != "Crunchwrap successfully smuggled into the parking lot." {
		t.Fatalf("expected recent Jump evidence caption, got %#v", recent.Evidence)
	}
}

func TestSubmissionWindowClosesAfterDeadline(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	future := time.Now().Add(1 * time.Hour)
	farFuture := time.Now().Add(2 * time.Hour)
	season := startSeasonWithDeadlines(t, server, "alice-token", group.Group.ID, future, farFuture)

	idea := createIdea(t, server, "alice-token", group.Group.ID, "Waffle House", "movie theater", "hash browns")
	plan := createPlannedJump(t, server, "alice-token", idea.ID, false)

	store.SetSeasonStatus(season.ActiveSeason.ID, "Judging Grace Period")

	auth := authorizeEvidenceUpload(t, server, "alice-token", plan.ID, "image/jpeg")
	submitRec := doJSON(server, http.MethodPost, "/v1/jumps/"+plan.ID+"/evidence", "alice-token", map[string]string{
		"uploadAuthorizationID": auth.ID,
		"caption":               "Attempted after deadline",
	})
	if submitRec.Code != http.StatusConflict {
		t.Fatalf("expected submission window closed status 409, got %d: %s", submitRec.Code, submitRec.Body.String())
	}
}

func TestGroupMemberCanJudgeAnotherPlayersPerformedJump(t *testing.T) {
	server, store := newGroupsTestServerAndStore(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
		"commitment":    4,
		"transgression": 3,
		"creativity":    3,
		"presentation":  2,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected Judgment status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var judgment judgmentBody
	decodeResponse(t, rec, &judgment)
	if judgment.ID == "" || judgment.JumpID != performed.Jump.ID {
		t.Fatalf("expected Judgment for Performed Jump, got %#v", judgment)
	}
	if judgment.PlayerID == performed.Jump.PlayerID {
		t.Fatalf("expected Judge to be a different Player than performer, got %#v", judgment)
	}
	if judgment.Commitment != 4 || judgment.Transgression != 3 || judgment.Creativity != 3 || judgment.Presentation != 2 {
		t.Fatalf("expected four submitted Judgment scores, got %#v", judgment)
	}
}

func TestGroupMemberCanRaiseDisputeOnVisiblePerformedJump(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before raising Dispute, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/disputes", "bob-token", map[string]string{
		"concern": "House Rules",
		"details": "This looked like it blocked the emergency exit.",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected Dispute status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var dispute disputeBody
	decodeResponse(t, rec, &dispute)
	if dispute.ID == "" || dispute.JumpID != performed.Jump.ID {
		t.Fatalf("expected created Dispute for visible Jump, got %#v", dispute)
	}
	if dispute.RaisedByPlayerID == performed.Jump.PlayerID {
		t.Fatalf("expected different Group member to raise Dispute, got %#v", dispute)
	}
	if dispute.Concern != "House Rules" || dispute.Details != "This looked like it blocked the emergency exit." || dispute.Status != "Open" {
		t.Fatalf("expected open House Rules Dispute, got %#v", dispute)
	}

	home := getGroupHome(t, server, "alice-token", group.Group.ID)
	if len(home.RecentJumps) != 1 || len(home.RecentJumps[0].Disputes) != 1 {
		t.Fatalf("expected Group home to show raised Dispute, got %#v", home.RecentJumps)
	}
	if home.RecentJumps[0].Disputes[0].ID != dispute.ID {
		t.Fatalf("expected Group home to show created Dispute, got %#v", home.RecentJumps[0].Disputes)
	}
}

func TestGroupMemberCanRaiseDisputeForEachMVPAcceptedConcern(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before raising Disputes, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	for _, concern := range []string{"House Rules", "Credibility", "Source", "Destination", "Food", "duplicate", "other"} {
		t.Run(concern, func(t *testing.T) {
			rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/disputes", "bob-token", map[string]string{
				"concern": concern,
				"details": "MVP concern coverage",
			})
			if rec.Code != http.StatusCreated {
				t.Fatalf("expected concern %q to be accepted, got %d: %s", concern, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestPerformerCannotJudgeTheirOwnPerformedJump(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	performed := performJump(t, server, "alice-token", group.Group.ID)

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "alice-token", map[string]int{
		"commitment":    4,
		"transgression": 3,
		"creativity":    3,
		"presentation":  2,
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected performer Judgment status 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJudgeCanEditTheirOneJudgmentWhileJudgingWindowIsOpen(t *testing.T) {
	server, store := newGroupsTestServerAndStore(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	created := submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 3, 2, http.StatusCreated)
	updated := submitJudgment(t, server, "bob-token", performed.Jump.ID, 2, 3, 4, 4, http.StatusOK)

	if updated.ID != created.ID || updated.JumpID != created.JumpID || updated.PlayerID != created.PlayerID {
		t.Fatalf("expected edited Judgment to keep identity, created %#v updated %#v", created, updated)
	}
	if updated.Commitment != 2 || updated.Transgression != 3 || updated.Creativity != 4 || updated.Presentation != 4 {
		t.Fatalf("expected edited Judgment scores, got %#v", updated)
	}
}

func TestJudgmentScoresMustStayInRange(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
		"commitment":    11,
		"transgression": 3,
		"creativity":    3,
		"presentation":  2,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected out-of-range Judgment score status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJudgmentRequiresAllScoreFields(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
		"commitment":    0,
		"transgression": 3,
		"creativity":    3,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected missing Judgment score status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJudgmentsCannotBeCreatedOrEditedAfterJudgingWindowCloses(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	season := startSeason(t, server, "alice-token", group.Group.ID)
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time { return time.Now().Add(11 * time.Minute) })

	created := submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 3, 2, http.StatusCreated)
	store.SetSeasonStatus(season.ActiveSeason.ID, "Finalized")

	editRec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
		"commitment":    2,
		"transgression": 3,
		"creativity":    4,
		"presentation":  4,
	})
	if editRec.Code != http.StatusConflict {
		t.Fatalf("expected closed Judging Window edit status 409, got %d: %s", editRec.Code, editRec.Body.String())
	}

	carolAcceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+createInvite(t, server, "alice-token", group.Group.ID).Token+"/accept", "carol-token", nil)
	if carolAcceptRec.Code != http.StatusOK {
		t.Fatalf("expected Carol to join Group before judging, got %d: %s", carolAcceptRec.Code, carolAcceptRec.Body.String())
	}
	createRec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "carol-token", map[string]int{
		"commitment":    1,
		"transgression": 1,
		"creativity":    1,
		"presentation":  1,
	})
	if createRec.Code != http.StatusConflict {
		t.Fatalf("expected closed Judging Window create status 409, got %d: %s", createRec.Code, createRec.Body.String())
	}

	if created.Commitment != 4 || created.Transgression != 3 || created.Creativity != 3 || created.Presentation != 2 {
		t.Fatalf("expected original Judgment to exist before closed-window attempts, got %#v", created)
	}
}

func TestFinalizedSeasonScoresJudgedJumpsAndLocksStandings(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	})
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	)
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 11, 0, 0, time.UTC)
	})

	submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 3, 2, http.StatusCreated)

	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	})

	home := getGroupHome(t, server, "alice-token", group.Group.ID)
	if home.ActiveSeason != nil {
		t.Fatalf("expected Finalized Season to no longer be active, got %#v", home.ActiveSeason)
	}
	if len(home.RecentJumps) != 1 {
		t.Fatalf("expected judged Jump to remain visible, got %#v", home.RecentJumps)
	}
	if home.RecentJumps[0].Jump.Status != "Judged Jump" || home.RecentJumps[0].Jump.FinalScore == nil || *home.RecentJumps[0].Jump.FinalScore != 12 {
		t.Fatalf("expected closed Performed Jump to receive Final Score 12, got %#v", home.RecentJumps[0].Jump)
	}
	if len(home.Standings) != 1 {
		t.Fatalf("expected one Standing entry, got %#v", home.Standings)
	}
	if home.Standings[0].Player.ID != group.Membership.PlayerID || home.Standings[0].SeasonScore != 12 || home.Standings[0].JudgedJumps != 1 {
		t.Fatalf("expected Alice to lead with Season Score 12 from one Judged Jump, got %#v", home.Standings[0])
	}

	editRec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
		"commitment":    4,
		"transgression": 4,
		"creativity":    4,
		"presentation":  4,
	})
	if editRec.Code != http.StatusConflict {
		t.Fatalf("expected Finalized Season to lock Judgment edits, got %d: %s", editRec.Code, editRec.Body.String())
	}
}

func TestSeasonCommissionerCanDisqualifySeasonLinkedJumpAndExcludeItFromStandings(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	})
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	)
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 11, 0, 0, time.UTC)
	})

	submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 3, 2, http.StatusCreated)

	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	})
	before := getGroupHome(t, server, "alice-token", group.Group.ID)
	if len(before.Standings) != 1 {
		t.Fatalf("expected judged Jump to reach Standings before Dispute resolution, got %#v", before.Standings)
	}

	dispute := raiseDispute(t, server, "bob-token", performed.Jump.ID, "Credibility", "The receipt timestamp does not match the Caption.")
	resolutionRec := doJSON(server, http.MethodPost, "/v1/disputes/"+dispute.ID+"/resolution", "alice-token", map[string]string{
		"resolution":       "Disqualified Jump",
		"resolutionReason": "Evidence does not support the claimed performance.",
	})
	if resolutionRec.Code != http.StatusOK {
		t.Fatalf("expected Dispute resolution status 200, got %d: %s", resolutionRec.Code, resolutionRec.Body.String())
	}

	var resolution disputeResolutionBody
	decodeResponse(t, resolutionRec, &resolution)
	if resolution.Jump.Status != "Disqualified Jump" {
		t.Fatalf("expected resolved Jump to become Disqualified Jump, got %#v", resolution)
	}
	if resolution.Dispute.Status != "Resolved" || resolution.Dispute.Resolution == nil || *resolution.Dispute.Resolution != "Disqualified Jump" {
		t.Fatalf("expected resolved Dispute to record disqualification, got %#v", resolution.Dispute)
	}

	after := getGroupHome(t, server, "bob-token", group.Group.ID)
	if len(after.RecentJumps) != 1 || after.RecentJumps[0].Jump.Status != "Disqualified Jump" {
		t.Fatalf("expected Disqualified Jump to remain visible, got %#v", after.RecentJumps)
	}
	if len(after.Standings) != 0 {
		t.Fatalf("expected Disqualified Jump to be excluded from Standings, got %#v", after.Standings)
	}
	if after.RecentJumps[0].Disputes[0].ResolutionReason == nil || *after.RecentJumps[0].Disputes[0].ResolutionReason != "Evidence does not support the claimed performance." {
		t.Fatalf("expected visible disqualification reason, got %#v", after.RecentJumps[0].Disputes)
	}
}

func TestFinalizedSeasonMarksUnjudgedJumpsAndExcludesOffSeasonJumpsFromStandings(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	})
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	)
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}

	unjudged := performJump(t, server, "alice-token", group.Group.ID)
	offSeasonIdea := createIdea(t, server, "alice-token", group.Group.ID, "Pizza Hut", "library", "personal pan pizza")
	offSeasonPlanned := createPlannedJump(t, server, "alice-token", offSeasonIdea.ID, true)
	offSeasonAuthorization := authorizeEvidenceUpload(t, server, "alice-token", offSeasonPlanned.ID, "image/jpeg")
	offSeasonRec := doJSON(server, http.MethodPost, "/v1/jumps/"+offSeasonPlanned.ID+"/evidence", "alice-token", map[string]string{
		"uploadAuthorizationId": offSeasonAuthorization.ID,
		"caption":               "Pizza Hut did make it to the library.",
	})
	if offSeasonRec.Code != http.StatusCreated {
		t.Fatalf("expected Off-Season Evidence status 201, got %d: %s", offSeasonRec.Code, offSeasonRec.Body.String())
	}

	// Advance past the Author Grace Period before judging
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 11, 0, 0, time.UTC)
	})

	submitJudgment(t, server, "bob-token", offSeasonPlanned.ID, 4, 4, 4, 4, http.StatusCreated)

	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	})

	home := getGroupHome(t, server, "alice-token", group.Group.ID)
	jumpsByID := map[string]jumpBody{}
	for _, recent := range home.RecentJumps {
		jumpsByID[recent.Jump.ID] = recent.Jump
	}
	if jumpsByID[unjudged.Jump.ID].Status != "Unjudged Jump" || jumpsByID[unjudged.Jump.ID].FinalScore != nil {
		t.Fatalf("expected Season-linked Jump without Judgments to be Unjudged, got %#v", jumpsByID[unjudged.Jump.ID])
	}
	if jumpsByID[offSeasonPlanned.ID].FinalScore != nil {
		t.Fatalf("expected Off-Season Jump to have no FinalScore from Season finalization, got %#v", jumpsByID[offSeasonPlanned.ID])
	}
	if jumpsByID[offSeasonPlanned.ID].Status != "Judged Jump" {
		t.Fatalf("expected Off-Season Jump to be Judged Jump after a Judgment, got %#v", jumpsByID[offSeasonPlanned.ID])
	}
	if len(home.Standings) != 0 {
		t.Fatalf("expected Off-Season Judgment and Unjudged Jump not to affect Standings, got %#v", home.Standings)
	}
}

func TestGroupAdminCanOverrideDisputeResolutionAndRemoveJumpFromNormalVisibility(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	})
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "carol-token", "Breakfast Crew")
	aliceInvite := createInvite(t, server, "carol-token", group.Group.ID)
	aliceAcceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+aliceInvite.Token+"/accept", "alice-token", nil)
	if aliceAcceptRec.Code != http.StatusOK {
		t.Fatalf("expected Alice to join Group before starting Season, got %d: %s", aliceAcceptRec.Code, aliceAcceptRec.Body.String())
	}
	bobInvite := createInvite(t, server, "carol-token", group.Group.ID)
	bobAcceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+bobInvite.Token+"/accept", "bob-token", nil)
	if bobAcceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", bobAcceptRec.Code, bobAcceptRec.Body.String())
	}
	startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
	)
	performed := performJump(t, server, "alice-token", group.Group.ID)
	dispute := raiseDispute(t, server, "bob-token", performed.Jump.ID, "other", "Identifiable bystander appears in the photo.")

	resolveRec := doJSON(server, http.MethodPost, "/v1/disputes/"+dispute.ID+"/resolution", "alice-token", map[string]string{
		"resolution":       "Disqualified Jump",
		"resolutionReason": "Competition evidence should be redacted before scoring.",
	})
	if resolveRec.Code != http.StatusOK {
		t.Fatalf("expected commissioner resolution status 200, got %d: %s", resolveRec.Code, resolveRec.Body.String())
	}

	overrideRec := doJSON(server, http.MethodPost, "/v1/disputes/"+dispute.ID+"/resolution", "carol-token", map[string]string{
		"resolution":       "Removed Jump",
		"resolutionReason": "Serious privacy violation requires hiding the Jump.",
	})
	if overrideRec.Code != http.StatusOK {
		t.Fatalf("expected Group Admin override status 200, got %d: %s", overrideRec.Code, overrideRec.Body.String())
	}

	var override disputeResolutionBody
	decodeResponse(t, overrideRec, &override)
	if override.Jump.Status != "Removed Jump" {
		t.Fatalf("expected Group Admin override to remove Jump, got %#v", override)
	}
	if override.Dispute.Status != "Overridden" || override.Dispute.OverrideResolution == nil || *override.Dispute.OverrideResolution != "Removed Jump" {
		t.Fatalf("expected override to remain visible on Dispute, got %#v", override.Dispute)
	}

	home := getGroupHome(t, server, "bob-token", group.Group.ID)
	if len(home.RecentJumps) != 0 {
		t.Fatalf("expected Removed Jump hidden from normal Group visibility, got %#v", home.RecentJumps)
	}
}

func TestGroupAdminCanResolveOpenOffSeasonDisputeAndRemoveJump(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before raising Dispute, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)
	if performed.Jump.SeasonID != nil || !performed.Jump.OffSeason {
		t.Fatalf("expected Off-Season Jump, got %#v", performed.Jump)
	}
	dispute := raiseDispute(t, server, "bob-token", performed.Jump.ID, "House Rules", "This should be removed from normal Group visibility.")

	resolutionRec := doJSON(server, http.MethodPost, "/v1/disputes/"+dispute.ID+"/resolution", "alice-token", map[string]string{
		"resolution":       "Removed Jump",
		"resolutionReason": "Off-Season Jump contains a serious privacy issue.",
	})
	if resolutionRec.Code != http.StatusOK {
		t.Fatalf("expected Group Admin off-season resolution status 200, got %d: %s", resolutionRec.Code, resolutionRec.Body.String())
	}

	var resolution disputeResolutionBody
	decodeResponse(t, resolutionRec, &resolution)
	if resolution.Jump.Status != "Removed Jump" {
		t.Fatalf("expected Group Admin to remove Off-Season Jump, got %#v", resolution)
	}
	if resolution.Dispute.Status != "Resolved" || resolution.Dispute.Resolution == nil || *resolution.Dispute.Resolution != "Removed Jump" {
		t.Fatalf("expected Off-Season Dispute to have terminal resolution, got %#v", resolution.Dispute)
	}

	home := getGroupHome(t, server, "bob-token", group.Group.ID)
	if len(home.RecentJumps) != 0 {
		t.Fatalf("expected Removed Off-Season Jump hidden from normal Group visibility, got %#v", home.RecentJumps)
	}
}

func TestStandingsUseLatestCreatedSeasonRatherThanSeasonIDOrder(t *testing.T) {
	currentTime := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
		return currentTime
	})
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}

	judgmentScores := [][4]int{
		{4, 4, 4, 4},
		{1, 1, 1, 1},
	}

	store.SetClock(func() time.Time { return currentTime })
	startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		currentTime.Add(time.Hour),
		currentTime.Add(2*time.Hour),
	)
	firstPerformed := performJump(t, server, "alice-token", group.Group.ID)

	store.SetClock(func() time.Time { return currentTime.Add(11 * time.Minute) })
	submitJudgment(t, server, "bob-token", firstPerformed.Jump.ID, judgmentScores[0][0], judgmentScores[0][1], judgmentScores[0][2], judgmentScores[0][3], http.StatusCreated)
	currentTime = currentTime.Add(3 * time.Hour)

	store.SetClock(func() time.Time { return currentTime })
	startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		currentTime.Add(time.Hour),
		currentTime.Add(2*time.Hour),
	)
	secondPerformed := performJump(t, server, "alice-token", group.Group.ID)

	store.SetClock(func() time.Time { return currentTime.Add(11 * time.Minute) })
	submitJudgment(t, server, "bob-token", secondPerformed.Jump.ID, judgmentScores[1][0], judgmentScores[1][1], judgmentScores[1][2], judgmentScores[1][3], http.StatusCreated)
	currentTime = currentTime.Add(3 * time.Hour)

	home := getGroupHome(t, server, "alice-token", group.Group.ID)
	if len(home.Standings) != 1 {
		t.Fatalf("expected one Standing entry, got %#v", home.Standings)
	}
	expectedSeasonScore := judgmentScores[1][0] + judgmentScores[1][1] + judgmentScores[1][2] + judgmentScores[1][3]
	if home.Standings[0].SeasonScore != expectedSeasonScore || home.Standings[0].JudgedJumps != 1 {
		t.Fatalf("expected latest season score %d to drive standings, got %#v", expectedSeasonScore, home.Standings[0])
	}
}

func TestIdeaAndPlannedJumpRequireGroupMembership(t *testing.T) {
	server := newGroupsTestServer(t)
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

	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-jump", "bob-token", nil)
	if planRec.Code != http.StatusForbidden {
		t.Fatalf("expected non-member Planned Jump status 403, got %d: %s", planRec.Code, planRec.Body.String())
	}
}

func TestGroupCannotStartSecondSeasonWhileActiveSeasonExists(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	firstRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", map[string]string{
		"submissionDeadline": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
		"judgingDeadline":    time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
	})
	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	secondRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", map[string]string{
		"submissionDeadline": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
		"judgingDeadline":    time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
	})
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

	secondRec := doJSON(restartedServer, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", map[string]string{
		"submissionDeadline": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
		"judgingDeadline":    time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
	})
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
			rec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", map[string]string{
				"submissionDeadline": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
				"judgingDeadline":    time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
			})
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

func TestPostgresAcceptInviteReturnsStandingsForFinalizedSeason(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	season := startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		time.Now().Add(24*time.Hour),
		time.Now().Add(48*time.Hour),
	)
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Postgres database: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `UPDATE jumps SET grace_period_expires_at = now() - interval '1 minute' WHERE id = $1`, performed.Jump.ID); err != nil {
		t.Fatalf("expire jump grace period: %v", err)
	}

	submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 3, 3, 2, http.StatusCreated)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close Postgres database: %v", err)
		}
	})
	if _, err := db.ExecContext(context.Background(), `
UPDATE seasons
SET submission_deadline = now() - interval '48 hours', judging_deadline = now() - interval '24 hours'
WHERE id = $1`, season.ActiveSeason.ID); err != nil {
		t.Fatalf("expire Season deadlines: %v", err)
	}

	carolAcceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+createInvite(t, server, "alice-token", group.Group.ID).Token+"/accept", "carol-token", nil)
	if carolAcceptRec.Code != http.StatusOK {
		t.Fatalf("expected Carol to join Group after judging, got %d: %s", carolAcceptRec.Code, carolAcceptRec.Body.String())
	}
	var accepted groupHomeBody
	decodeResponse(t, carolAcceptRec, &accepted)
	if len(accepted.Standings) != 1 {
		t.Fatalf("expected standings in Invite accept response, got %#v", accepted.Standings)
	}
	if accepted.Standings[0].Player.ID != group.Membership.PlayerID || accepted.Standings[0].SeasonScore != 12 || accepted.Standings[0].JudgedJumps != 1 {
		t.Fatalf("expected Alice standings in Invite accept response, got %#v", accepted.Standings[0])
	}
}

func TestPostgresGroupAdminEmergencySeasonOverridesAppearInSeasonHistory(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "bob-token", "Breakfast Crew")
	invite := createInvite(t, server, "bob-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "alice-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Alice to join Group before starting Season, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	season := startSeason(t, server, "alice-token", group.Group.ID)

	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "bob-token", nil)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("expected Group Admin override close status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
	}
	finalizeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/finalize", "bob-token", nil)
	if finalizeRec.Code != http.StatusOK {
		t.Fatalf("expected Group Admin override finalize status 200, got %d: %s", finalizeRec.Code, finalizeRec.Body.String())
	}

	history := getSeasonHistory(t, server, "alice-token", season.ActiveSeason.ID)
	if len(history.Entries) != 2 {
		t.Fatalf("expected two visible override actions in Season history, got %#v", history.Entries)
	}
	if history.Entries[0].Action != "Submissions Closed" || !history.Entries[0].Override || history.Entries[0].ActorRole != "Group Admin" {
		t.Fatalf("expected visible Group Admin close override entry, got %#v", history.Entries[0])
	}
	if history.Entries[1].Action != "Season Finalized" || !history.Entries[1].Override || history.Entries[1].ToStatus != "Finalized" {
		t.Fatalf("expected visible Group Admin finalize override entry, got %#v", history.Entries[1])
	}
}

func TestPostgresGroupAdminCanResolveOpenOffSeasonDisputeAndRemoveJump(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	invite := createInvite(t, server, "alice-token", group.Group.ID)
	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected Bob to join Group before raising Dispute, got %d: %s", acceptRec.Code, acceptRec.Body.String())
	}
	performed := performJump(t, server, "alice-token", group.Group.ID)
	if performed.Jump.SeasonID != nil || !performed.Jump.OffSeason {
		t.Fatalf("expected Off-Season Jump, got %#v", performed.Jump)
	}
	dispute := raiseDispute(t, server, "bob-token", performed.Jump.ID, "House Rules", "This should be removed from normal Group visibility.")

	resolutionRec := doJSON(server, http.MethodPost, "/v1/disputes/"+dispute.ID+"/resolution", "alice-token", map[string]string{
		"resolution":       "Removed Jump",
		"resolutionReason": "Off-Season Jump contains a serious privacy issue.",
	})
	if resolutionRec.Code != http.StatusOK {
		t.Fatalf("expected Group Admin off-season resolution status 200, got %d: %s", resolutionRec.Code, resolutionRec.Body.String())
	}

	var resolution disputeResolutionBody
	decodeResponse(t, resolutionRec, &resolution)
	if resolution.Jump.Status != "Removed Jump" {
		t.Fatalf("expected Group Admin to remove Off-Season Jump, got %#v", resolution)
	}
	if resolution.Dispute.Status != "Resolved" || resolution.Dispute.Resolution == nil || *resolution.Dispute.Resolution != "Removed Jump" {
		t.Fatalf("expected Off-Season Dispute to have terminal resolution, got %#v", resolution.Dispute)
	}

	home := getGroupHome(t, server, "bob-token", group.Group.ID)
	if len(home.RecentJumps) != 0 {
		t.Fatalf("expected Removed Off-Season Jump hidden from normal Group visibility, got %#v", home.RecentJumps)
	}
}

func TestPostgresCloseSubmissionsBlocksCompetitionEvidenceSubmission(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	season := startSeason(t, server, "alice-token", group.Group.ID)
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Waffle House", "movie theater", "hash browns")
	planned := createPlannedJump(t, server, "alice-token", idea.ID, false)
	authorization := authorizeEvidenceUpload(t, server, "alice-token", planned.ID, "image/jpeg")

	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "alice-token", nil)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
	}

	submitRec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence", "alice-token", map[string]string{
		"uploadAuthorizationId": authorization.ID,
		"caption":               "Too late for competition.",
	})
	if submitRec.Code != http.StatusConflict {
		t.Fatalf("expected closed submission status 409, got %d: %s", submitRec.Code, submitRec.Body.String())
	}
}

func TestPostgresConcurrentIdeaCreationReturnsDistinctIdeas(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Concurrent Idea Creators")

	const ideaCount = 4
	ideas := make(chan jumpBody, ideaCount)
	errors := make(chan string, ideaCount)
	var wg sync.WaitGroup
	for i := 0; i < ideaCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/ideas", "alice-token", map[string]string{
				"source":      "Taco Bell",
				"destination": "Olive Garden parking lot",
				"food":        "Crunchwrap",
			})
			if rec.Code != http.StatusCreated {
				errors <- rec.Body.String()
				return
			}
			var idea jumpBody
			if err := json.NewDecoder(rec.Body).Decode(&idea); err != nil {
				errors <- err.Error()
				return
			}
			ideas <- idea
		}()
	}
	wg.Wait()
	close(ideas)
	close(errors)
	for err := range errors {
		t.Fatalf("expected concurrent Idea creation to succeed, got %s", err)
	}

	ids := []string{}
	for idea := range ideas {
		ids = append(ids, idea.ID)
	}
	if len(ids) != ideaCount {
		t.Fatalf("expected %d Ideas, got %d", ideaCount, len(ids))
	}
	sort.Strings(ids)
	for i := 1; i < len(ids); i++ {
		if ids[i] == ids[i-1] {
			t.Fatalf("expected distinct Idea ids, got %v", ids)
		}
	}
}

func TestPostgresConcurrentPlannedJumpCreationOnlyTransitionsIdeaOnce(t *testing.T) {
	databaseURL := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set SUPPERJUMPIN_TEST_DATABASE_URL to run durable Postgres behavior test")
	}
	installSlowJumpUpdateTrigger(t, databaseURL)

	store := newPostgresTestStore(t, databaseURL)
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Concurrent Planners")
	startSeason(t, server, "alice-token", group.Group.ID)
	idea := createIdea(t, server, "alice-token", group.Group.ID, "Waffle House", "movie theater", "hash browns")

	codes := make(chan int, 2)
	var wg sync.WaitGroup
	for _, body := range []any{nil, map[string]bool{"offSeason": true}} {
		wg.Add(1)
		go func(body any) {
			defer wg.Done()
			rec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-jump", "alice-token", body)
			codes <- rec.Code
		}(body)
	}
	wg.Wait()
	close(codes)

	seen := map[int]int{}
	for code := range codes {
		seen[code]++
	}
	if seen[http.StatusCreated] != 1 || seen[http.StatusNotFound] != 1 {
		t.Fatalf("expected one Planned Jump creation and one rejected transition, got %#v", seen)
	}
}

func TestCreatePerformedJumpWithoutGroupIdIsOffSeason(t *testing.T) {
	server := newGroupsTestServer(t)
	createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodPost, "/v1/jumps", "alice-token", map[string]any{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           "Crunchwrap",
		"caption":        "Crunchwrap successfully smuggled into the parking lot.",
		"mediaObjectKey": "uploads/test/1",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var jump jumpBody
	decodeResponse(t, rec, &jump)
	if jump.SeasonID != nil {
		t.Fatalf("expected null SeasonID when no groupId, got %v", *jump.SeasonID)
	}
	if !jump.OffSeason {
		t.Fatalf("expected OffSeason true when no groupId")
	}
	if jump.GroupID != "" {
		t.Fatalf("expected empty GroupID when no groupId, got %q", jump.GroupID)
	}
}

func TestCreatePerformedJumpGracePeriodUsesInjectableClock(t *testing.T) {
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
		return now
	})
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodPost, "/v1/jumps", "alice-token", map[string]any{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           "Crunchwrap",
		"caption":        "Crunchwrap successfully smuggled into the parking lot.",
		"mediaObjectKey": "uploads/test/1",
		"groupId":        group.Group.ID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var jump jumpBody
	decodeResponse(t, rec, &jump)
	expectedGrace := now.Add(10 * time.Minute)
	if !jump.GracePeriodExpiresAt.Equal(expectedGrace) {
		t.Fatalf("expected gracePeriodExpiresAt %v, got %v", expectedGrace, jump.GracePeriodExpiresAt)
	}

	store.SetClock(func() time.Time {
		return now.Add(1 * time.Hour)
	})
	rec2 := doJSON(server, http.MethodPost, "/v1/jumps", "alice-token", map[string]any{
		"source":         "Waffle House",
		"destination":    "movie theater",
		"food":           "hash browns",
		"caption":        "Hash browns at the movies.",
		"mediaObjectKey": "uploads/test/2",
		"groupId":        group.Group.ID,
	})
	if rec2.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var jump2 jumpBody
	decodeResponse(t, rec2, &jump2)
	expectedGrace2 := now.Add(1*time.Hour + 10*time.Minute)
	if !jump2.GracePeriodExpiresAt.Equal(expectedGrace2) {
		t.Fatalf("expected gracePeriodExpiresAt %v, got %v", expectedGrace2, jump2.GracePeriodExpiresAt)
	}
}

func TestCreatePerformedJumpLinksToActiveSeason(t *testing.T) {
	store := newCleanPostgresTestStore(t)
	store.SetClock(func() time.Time {
		return time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	})
	server := newGroupsTestServerWithStore(store)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")
	season := startSeasonWithDeadlines(
		t,
		server,
		"alice-token",
		group.Group.ID,
		time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
	)

	rec := doJSON(server, http.MethodPost, "/v1/jumps", "alice-token", map[string]any{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           "Crunchwrap",
		"caption":        "Crunchwrap successfully smuggled into the parking lot.",
		"mediaObjectKey": "uploads/test/1",
		"groupId":        group.Group.ID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var jump jumpBody
	decodeResponse(t, rec, &jump)
	if jump.SeasonID == nil || *jump.SeasonID != season.ActiveSeason.ID {
		t.Fatalf("expected jump to be linked to active Season %q, got %v", season.ActiveSeason.ID, jump.SeasonID)
	}
	if jump.OffSeason {
		t.Fatalf("expected OffSeason false when linked to active season")
	}
}

func TestCreatePerformedJumpAppearsInGroupHome(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodPost, "/v1/jumps", "alice-token", map[string]any{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           "Crunchwrap",
		"caption":        "Crunchwrap successfully smuggled into the parking lot.",
		"mediaObjectKey": "uploads/test/1",
		"groupId":        group.Group.ID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var createdJump jumpBody
	decodeResponse(t, rec, &createdJump)

	home := getGroupHome(t, server, "alice-token", group.Group.ID)
	if len(home.RecentJumps) != 1 {
		t.Fatalf("expected one recent Jump on Group home, got %d", len(home.RecentJumps))
	}
	recent := home.RecentJumps[0]
	if recent.Jump.ID != createdJump.ID {
		t.Fatalf("expected recent Jump to match created Jump, got %#v", recent)
	}
	if recent.Jump.Status != "Performed Jump" {
		t.Fatalf("expected recent Jump status 'Performed Jump', got %q", recent.Jump.Status)
	}
	if recent.Evidence.MediaObjectKey != "uploads/test/1" {
		t.Fatalf("expected evidence mediaObjectKey on recent Jump, got %q", recent.Evidence.MediaObjectKey)
	}
}

func TestCreatePerformedJumpThenSelfJudgmentBlocked(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodPost, "/v1/jumps", "alice-token", map[string]any{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           "Crunchwrap",
		"caption":        "Crunchwrap successfully smuggled into the parking lot.",
		"mediaObjectKey": "uploads/test/1",
		"groupId":        group.Group.ID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var createdJump jumpBody
	decodeResponse(t, rec, &createdJump)

	judgeRec := doJSON(server, http.MethodPost, "/v1/jumps/"+createdJump.ID+"/judgment", "alice-token", map[string]int{
		"commitment":    4,
		"transgression": 3,
		"creativity":    3,
		"presentation":  2,
	})
	if judgeRec.Code != http.StatusForbidden {
		t.Fatalf("expected self-judgment status 403, got %d: %s", judgeRec.Code, judgeRec.Body.String())
	}
}

func TestCreatePerformedJumpRejectsUnauthenticated(t *testing.T) {
	server := newGroupsTestServer(t)

	rec := doJSON(server, http.MethodPost, "/v1/jumps", "", map[string]any{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           "Crunchwrap",
		"caption":        "Crunchwrap successfully smuggled into the parking lot.",
		"mediaObjectKey": "uploads/test/1",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreatePerformedJumpRejectsMissingFields(t *testing.T) {
	server := newGroupsTestServer(t)

	rec := doJSON(server, http.MethodPost, "/v1/jumps", "alice-token", map[string]any{
		"source":  "Taco Bell",
		"food":    "Crunchwrap",
		"caption": "Test caption",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreatePerformedJumpDirectly(t *testing.T) {
	server := newGroupsTestServer(t)
	group := createGroup(t, server, "alice-token", "Breakfast Crew")

	rec := doJSON(server, http.MethodPost, "/v1/jumps", "alice-token", map[string]any{
		"source":         "Taco Bell",
		"destination":    "Olive Garden parking lot",
		"food":           "Crunchwrap",
		"caption":        "Crunchwrap successfully smuggled into the parking lot.",
		"mediaObjectKey": "uploads/test/1",
		"groupId":        group.Group.ID,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var jump jumpBody
	decodeResponse(t, rec, &jump)
	if jump.ID == "" {
		t.Fatalf("expected created Jump to have an id")
	}
	if jump.Status != "Performed Jump" {
		t.Fatalf("expected Jump status 'Performed Jump', got %q", jump.Status)
	}
	if jump.Source != "Taco Bell" || jump.Destination != "Olive Garden parking lot" || jump.Food != "Crunchwrap" {
		t.Fatalf("expected Source, Destination, and Food from request, got %#v", jump)
	}
	if jump.GroupID != group.Group.ID {
		t.Fatalf("expected Jump GroupID %q, got %q", group.Group.ID, jump.GroupID)
	}
	if jump.PlayerID == "" {
		t.Fatalf("expected Jump to have a PlayerID")
	}
	if !jump.OffSeason {
		t.Fatalf("expected OffSeason true when group has no active season")
	}
	if jump.SeasonID != nil {
		t.Fatalf("expected no SeasonID when no active season, got %v", *jump.SeasonID)
	}
	if jump.GracePeriodExpiresAt.IsZero() {
		t.Fatalf("expected gracePeriodExpiresAt to be set")
	}
}

func newGroupsTestServer(t *testing.T) http.Handler {
	return newGroupsTestServerWithPersistence(newCleanPostgresTestStore(t))
}

func newGroupsTestServerAndStore(t *testing.T) (http.Handler, *httpapi.PostgresStore) {
	store := newCleanPostgresTestStore(t)
	return newGroupsTestServerWithPersistence(store), store
}

func newGroupsTestServerWithStore(store httpapi.Store) http.Handler {
	return newGroupsTestServerWithPersistence(store.(httpapi.Persistence))
}

func newGroupsTestServerWithPersistence(db httpapi.Persistence) http.Handler {
	return httpapi.NewServer(httpapi.ServerConfig{
		Auth: httpapi.StaticAuthVerifier{
			"alice-token": {Provider: "supabase", Subject: "alice-auth", Email: "alice@example.com"},
			"bob-token":   {Provider: "supabase", Subject: "bob-auth", Email: "bob@example.com"},
			"carol-token": {Provider: "supabase", Subject: "carol-auth", Email: "carol@example.com"},
		},
		Store: db.(httpapi.Store),
		DB:    db,
	})
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

func installSlowJumpUpdateTrigger(t *testing.T, databaseURL string) {
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
CREATE OR REPLACE FUNCTION supperjumpin_test_slow_jump_update()
RETURNS trigger AS $$
BEGIN
  PERFORM pg_sleep(0.2);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS supperjumpin_test_slow_jump_update ON jumps;
CREATE TRIGGER supperjumpin_test_slow_jump_update
BEFORE UPDATE ON jumps
FOR EACH ROW EXECUTE FUNCTION supperjumpin_test_slow_jump_update();`); err != nil {
		t.Fatalf("install slow Jump update trigger: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `
DROP TRIGGER IF EXISTS supperjumpin_test_slow_jump_update ON jumps;
DROP FUNCTION IF EXISTS supperjumpin_test_slow_jump_update();`); err != nil {
			t.Fatalf("remove slow Jump update trigger: %v", err)
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
	RecentJumps []performedJumpViewBody `json:"recentJumps"`
	Standings   []standingBody          `json:"standings"`
}

type performedJumpViewBody struct {
	Jump struct {
		ID                   string    `json:"id"`
		GroupID              string    `json:"groupId"`
		PlayerID             string    `json:"playerId"`
		SeasonID             *string   `json:"seasonId"`
		Status               string    `json:"status"`
		Source               string    `json:"source"`
		Destination          string    `json:"destination"`
		Food                 string    `json:"food"`
		OffSeason            bool      `json:"offSeason"`
		FinalScore           *int      `json:"finalScore"`
		GracePeriodExpiresAt time.Time `json:"gracePeriodExpiresAt"`
	} `json:"jump"`
	Performer struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	} `json:"performer"`
	Evidence struct {
		ID             string `json:"id"`
		JumpID         string `json:"jumpId"`
		Caption        string `json:"caption"`
		MediaObjectKey string `json:"mediaObjectKey"`
	} `json:"evidence"`
	Disputes []disputeBody `json:"disputes"`
}

type disputeBody struct {
	ID                 string  `json:"id"`
	JumpID             string  `json:"jumpId"`
	RaisedByPlayerID   string  `json:"raisedByPlayerId"`
	Concern            string  `json:"concern"`
	Details            string  `json:"details"`
	Status             string  `json:"status"`
	Resolution         *string `json:"resolution"`
	ResolutionReason   *string `json:"resolutionReason"`
	ResolvedByPlayerID *string `json:"resolvedByPlayerId"`
	OverrideResolution *string `json:"overrideResolution"`
	OverrideReason     *string `json:"overrideReason"`
	OverrideByPlayerID *string `json:"overrideByPlayerId"`
}

type disputeResolutionBody struct {
	Jump    jumpBody    `json:"jump"`
	Dispute disputeBody `json:"dispute"`
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
	rec := doJSON(server, http.MethodPost, "/v1/groups/"+groupID+"/seasons", token, map[string]string{
		"submissionDeadline": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
		"judgingDeadline":    time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body groupHomeBody
	decodeResponse(t, rec, &body)
	return body
}

func startSeasonWithDeadlines(t *testing.T, server http.Handler, token string, groupID string, submissionDeadline time.Time, judgingDeadline time.Time) groupHomeBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/groups/"+groupID+"/seasons", token, map[string]string{
		"submissionDeadline": submissionDeadline.Format(time.RFC3339),
		"judgingDeadline":    judgingDeadline.Format(time.RFC3339),
	})
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
	normalizeGroupHomeBody(&body)
	return body
}

func getSeasonHistory(t *testing.T, server http.Handler, token string, seasonID string) seasonHistoryBody {
	t.Helper()
	rec := doJSON(server, http.MethodGet, "/v1/seasons/"+seasonID+"/history", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body seasonHistoryBody
	decodeResponse(t, rec, &body)
	return body
}

type inviteBody struct {
	ID      string `json:"id"`
	GroupID string `json:"groupId"`
	Token   string `json:"token"`
}

type jumpBody struct {
	ID                   string    `json:"id"`
	GroupID              string    `json:"groupId"`
	PlayerID             string    `json:"playerId"`
	SeasonID             *string   `json:"seasonId"`
	Status               string    `json:"status"`
	Source               string    `json:"source"`
	Destination          string    `json:"destination"`
	Food                 string    `json:"food"`
	OffSeason            bool      `json:"offSeason"`
	FinalScore           *int      `json:"finalScore"`
	GracePeriodExpiresAt time.Time `json:"gracePeriodExpiresAt"`
}

type standingBody struct {
	Player struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	} `json:"player"`
	SeasonScore int `json:"seasonScore"`
	JudgedJumps int `json:"judgedJumps"`
}

type evidenceUploadAuthorizationBody struct {
	ID             string            `json:"id"`
	JumpID         string            `json:"jumpId"`
	UploadURL      string            `json:"uploadUrl"`
	UploadMethod   string            `json:"uploadMethod"`
	UploadHeaders  map[string]string `json:"uploadHeaders"`
	MediaObjectKey string            `json:"mediaObjectKey"`
	ExpiresAt      string            `json:"expiresAt"`
}

type evidenceBody struct {
	ID             string `json:"id"`
	JumpID         string `json:"jumpId"`
	Caption        string `json:"caption"`
	MediaObjectKey string `json:"mediaObjectKey"`
}

type evidenceSubmissionBody struct {
	Jump     jumpBody     `json:"jump"`
	Evidence evidenceBody `json:"evidence"`
}

type judgmentBody struct {
	ID            string `json:"id"`
	JumpID        string `json:"jumpId"`
	PlayerID      string `json:"playerId"`
	Commitment    int    `json:"commitment"`
	Transgression int    `json:"transgression"`
	Creativity    int    `json:"creativity"`
	Presentation  int    `json:"presentation"`
}

type seasonHistoryBody struct {
	Entries []struct {
		Action        string `json:"action"`
		ActorPlayerID string `json:"actorPlayerId"`
		ActorRole     string `json:"actorRole"`
		Override      bool   `json:"override"`
		ToStatus      string `json:"toStatus"`
	} `json:"entries"`
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

func createIdea(t *testing.T, server http.Handler, token string, groupID string, source string, destination string, food string) jumpBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/groups/"+groupID+"/ideas", token, map[string]string{
		"source":      source,
		"destination": destination,
		"food":        food,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body jumpBody
	decodeResponse(t, rec, &body)
	return body
}

func createPlannedJump(t *testing.T, server http.Handler, token string, ideaID string, offSeason bool) jumpBody {
	t.Helper()
	var body any
	if offSeason {
		body = map[string]bool{"offSeason": true}
	}
	rec := doJSON(server, http.MethodPost, "/v1/ideas/"+ideaID+"/planned-jump", token, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var planned jumpBody
	decodeResponse(t, rec, &planned)
	return planned
}

func authorizeEvidenceUpload(t *testing.T, server http.Handler, token string, jumpID string, contentType string) evidenceUploadAuthorizationBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+jumpID+"/evidence-upload-authorizations", token, map[string]string{
		"contentType": contentType,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var authorization evidenceUploadAuthorizationBody
	decodeResponse(t, rec, &authorization)
	return authorization
}

func performJump(t *testing.T, server http.Handler, token string, groupID string) evidenceSubmissionBody {
	t.Helper()
	idea := createIdea(t, server, token, groupID, "Taco Bell", "Olive Garden parking lot", "Crunchwrap")
	planned := createPlannedJump(t, server, token, idea.ID, false)
	authorization := authorizeEvidenceUpload(t, server, token, planned.ID, "image/jpeg")
	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence", token, map[string]string{
		"uploadAuthorizationId": authorization.ID,
		"caption":               "Crunchwrap successfully smuggled into the parking lot.",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var submission evidenceSubmissionBody
	decodeResponse(t, rec, &submission)
	normalizeEvidenceSubmissionBody(&submission)
	return submission
}

func raiseDispute(t *testing.T, server http.Handler, token string, jumpID string, concern string, details string) disputeBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+jumpID+"/disputes", token, map[string]string{
		"concern": concern,
		"details": details,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected Dispute status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var dispute disputeBody
	decodeResponse(t, rec, &dispute)
	return dispute
}

func submitJudgment(t *testing.T, server http.Handler, token string, jumpID string, commitment int, transgression int, creativity int, presentation int, expectedStatus int) judgmentBody {
	t.Helper()
	rec := doJSON(server, http.MethodPost, "/v1/jumps/"+jumpID+"/judgment", token, map[string]int{
		"commitment":    commitment,
		"transgression": transgression,
		"creativity":    creativity,
		"presentation":  presentation,
	})
	if rec.Code != expectedStatus {
		t.Fatalf("expected Judgment status %d, got %d: %s", expectedStatus, rec.Code, rec.Body.String())
	}
	var judgment judgmentBody
	decodeResponse(t, rec, &judgment)
	return judgment
}

func normalizeGroupHomeBody(body *groupHomeBody) {}

func normalizeEvidenceSubmissionBody(body *evidenceSubmissionBody) {}
