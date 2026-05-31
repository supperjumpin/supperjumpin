     1|     1|     1|     1|package httpapi_test
     2|     2|     2|     2|
     3|     3|     3|     3|import (
     4|     4|     4|     4|	"bytes"
     5|     5|     5|     5|	"context"
     6|     6|     6|     6|	"database/sql"
     7|     7|     7|     7|	"encoding/json"
     8|     8|     8|     8|	"net/http"
     9|     9|     9|     9|	"net/http/httptest"
    10|    10|    10|    10|	"os"
    11|    11|    11|    11|	"sort"
    12|    12|    12|    12|	"sync"
    13|    13|    13|    13|	"testing"
    14|    14|    14|    14|	"time"
    15|    15|    15|    15|
    16|    16|    16|    16|	_ "github.com/jackc/pgx/v5/stdlib"
    17|    17|    17|    17|
    18|    18|    18|    18|	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
    19|    19|    19|    19|)
    20|    20|    20|    20|
    21|    21|    21|    21|func TestCreateGroupMakesSignedInPlayerGroupAdminAndReturnsGroupHome(t *testing.T) {
    22|    22|    22|    22|	server := newGroupsTestServer()
    23|    23|    23|    23|
    24|    24|    24|    24|	createRec := doJSON(server, http.MethodPost, "/v1/groups", "alice-token", map[string]string{"name": "Breakfast Crew"})
    25|    25|    25|    25|	if createRec.Code != http.StatusCreated {
    26|    26|    26|    26|		t.Fatalf("expected status 201, got %d: %s", createRec.Code, createRec.Body.String())
    27|    27|    27|    27|	}
    28|    28|    28|    28|
    29|    29|    29|    29|	var created struct {
    30|    30|    30|    30|		Group struct {
    31|    31|    31|    31|			ID   string `json:"id"`
    32|    32|    32|    32|			Name string `json:"name"`
    33|    33|    33|    33|		} `json:"group"`
    34|    34|    34|    34|		Membership struct {
    35|    35|    35|    35|			PlayerID string `json:"playerId"`
    36|    36|    36|    36|			Role     string `json:"role"`
    37|    37|    37|    37|		} `json:"membership"`
    38|    38|    38|    38|		ActiveSeason any   `json:"activeSeason"`
    39|    39|    39|    39|		RecentJumps []any `json:"recentJumps"`
    40|    40|    40|    40|		Standings    []any `json:"standings"`
    41|    41|    41|    41|	}
    42|    42|    42|    42|	decodeResponse(t, createRec, &created)
    43|    43|    43|    43|
    44|    44|    44|    44|	if created.Group.ID == "" {
    45|    45|    45|    45|		t.Fatalf("expected created Group to have an id")
    46|    46|    46|    46|	}
    47|    47|    47|    47|	if created.Group.Name != "Breakfast Crew" {
    48|    48|    48|    48|		t.Fatalf("expected Group name from request, got %q", created.Group.Name)
    49|    49|    49|    49|	}
    50|    50|    50|    50|	if created.Membership.PlayerID == "" {
    51|    51|    51|    51|		t.Fatalf("expected Group Membership to identify the Player")
    52|    52|    52|    52|	}
    53|    53|    53|    53|	if created.Membership.Role != "Group Admin" {
    54|    54|    54|    54|		t.Fatalf("expected creator to become Group Admin, got %q", created.Membership.Role)
    55|    55|    55|    55|	}
    56|    56|    56|    56|	if created.ActiveSeason != nil {
    57|    57|    57|    57|		t.Fatalf("expected no Active Season, got %#v", created.ActiveSeason)
    58|    58|    58|    58|	}
    59|    59|    59|    59|	if len(created.RecentJumps) != 0 {
    60|    60|    60|    60|		t.Fatalf("expected no recent Jumps, got %#v", created.RecentJumps)
    61|    61|    61|    61|	}
    62|    62|    62|    62|	if len(created.Standings) != 0 {
    63|    63|    63|    63|		t.Fatalf("expected empty Standings, got %#v", created.Standings)
    64|    64|    64|    64|	}
    65|    65|    65|    65|
    66|    66|    66|    66|	homeRec := doJSON(server, http.MethodGet, "/v1/groups/"+created.Group.ID+"/home", "alice-token", nil)
    67|    67|    67|    67|	if homeRec.Code != http.StatusOK {
    68|    68|    68|    68|		t.Fatalf("expected status 200, got %d: %s", homeRec.Code, homeRec.Body.String())
    69|    69|    69|    69|	}
    70|    70|    70|    70|	var home struct {
    71|    71|    71|    71|		Group struct {
    72|    72|    72|    72|			ID string `json:"id"`
    73|    73|    73|    73|		} `json:"group"`
    74|    74|    74|    74|		Membership struct {
    75|    75|    75|    75|			Role string `json:"role"`
    76|    76|    76|    76|		} `json:"membership"`
    77|    77|    77|    77|	}
    78|    78|    78|    78|	decodeResponse(t, homeRec, &home)
    79|    79|    79|    79|	if home.Group.ID != created.Group.ID || home.Membership.Role != "Group Admin" {
    80|    80|    80|    80|		t.Fatalf("expected backend Group home for created Group, got %#v", home)
    81|    81|    81|    81|	}
    82|    82|    82|    82|}
    83|    83|    83|    83|
    84|    84|    84|    84|func TestSignedInPlayerCanListMultipleGroupsAndSwitchGroupHome(t *testing.T) {
    85|    85|    85|    85|	server := newGroupsTestServer()
    86|    86|    86|    86|
    87|    87|    87|    87|	breakfast := createGroup(t, server, "alice-token", "Breakfast Crew")
    88|    88|    88|    88|	dinner := createGroup(t, server, "alice-token", "Dinner Weirdos")
    89|    89|    89|    89|
    90|    90|    90|    90|	listRec := doJSON(server, http.MethodGet, "/v1/groups", "alice-token", nil)
    91|    91|    91|    91|	if listRec.Code != http.StatusOK {
    92|    92|    92|    92|		t.Fatalf("expected status 200, got %d: %s", listRec.Code, listRec.Body.String())
    93|    93|    93|    93|	}
    94|    94|    94|    94|	var list struct {
    95|    95|    95|    95|		Memberships []struct {
    96|    96|    96|    96|			Group struct {
    97|    97|    97|    97|				ID   string `json:"id"`
    98|    98|    98|    98|				Name string `json:"name"`
    99|    99|    99|    99|			} `json:"group"`
   100|   100|   100|   100|			Membership struct {
   101|   101|   101|   101|				Role string `json:"role"`
   102|   102|   102|   102|			} `json:"membership"`
   103|   103|   103|   103|		} `json:"memberships"`
   104|   104|   104|   104|	}
   105|   105|   105|   105|	decodeResponse(t, listRec, &list)
   106|   106|   106|   106|	if len(list.Memberships) != 2 {
   107|   107|   107|   107|		t.Fatalf("expected two Group Memberships, got %#v", list.Memberships)
   108|   108|   108|   108|	}
   109|   109|   109|   109|
   110|   110|   110|   110|	breakfastHome := getGroupHome(t, server, "alice-token", breakfast.Group.ID)
   111|   111|   111|   111|	dinnerHome := getGroupHome(t, server, "alice-token", dinner.Group.ID)
   112|   112|   112|   112|	if breakfastHome.Group.Name != "Breakfast Crew" {
   113|   113|   113|   113|		t.Fatalf("expected switched Group home for Breakfast Crew, got %#v", breakfastHome.Group)
   114|   114|   114|   114|	}
   115|   115|   115|   115|	if dinnerHome.Group.Name != "Dinner Weirdos" {
   116|   116|   116|   116|		t.Fatalf("expected switched Group home for Dinner Weirdos, got %#v", dinnerHome.Group)
   117|   117|   117|   117|	}
   118|   118|   118|   118|}
   119|   119|   119|   119|
   120|   120|   120|   120|func TestGroupHomeRejectsSignedInNonMember(t *testing.T) {
   121|   121|   121|   121|	server := newGroupsTestServer()
   122|   122|   122|   122|	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
   123|   123|   123|   123|
   124|   124|   124|   124|	rec := doJSON(server, http.MethodGet, "/v1/groups/"+aliceGroup.Group.ID+"/home", "bob-token", nil)
   125|   125|   125|   125|	if rec.Code != http.StatusForbidden {
   126|   126|   126|   126|		t.Fatalf("expected status 403, got %d: %s", rec.Code, rec.Body.String())
   127|   127|   127|   127|	}
   128|   128|   128|   128|}
   129|   129|   129|   129|
   130|   130|   130|   130|func TestGroupMemberCreatesInviteAndSignedInPlayerAcceptsWithoutReplacingExistingPlayHistory(t *testing.T) {
   131|   131|   131|   131|	server := newGroupsTestServer()
   132|   132|   132|   132|	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
   133|   133|   133|   133|	bobExistingGroup := createGroup(t, server, "bob-token", "Dinner Weirdos")
   134|   134|   134|   134|
   135|   135|   135|   135|	invite := createInvite(t, server, "alice-token", aliceGroup.Group.ID)
   136|   136|   136|   136|	if invite.ID == "" || invite.GroupID != aliceGroup.Group.ID || invite.Token == "" {
   137|   137|   137|   137|		t.Fatalf("expected Invite for Alice's Group, got %#v", invite)
   138|   138|   138|   138|	}
   139|   139|   139|   139|
   140|   140|   140|   140|	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
   141|   141|   141|   141|	if acceptRec.Code != http.StatusOK {
   142|   142|   142|   142|		t.Fatalf("expected status 200, got %d: %s", acceptRec.Code, acceptRec.Body.String())
   143|   143|   143|   143|	}
   144|   144|   144|   144|	var accepted groupHomeBody
   145|   145|   145|   145|	decodeResponse(t, acceptRec, &accepted)
   146|   146|   146|   146|	if accepted.Group.ID != aliceGroup.Group.ID || accepted.Membership.Role != "Player" {
   147|   147|   147|   147|		t.Fatalf("expected Bob to join Alice's Group as Player, got %#v", accepted)
   148|   148|   148|   148|	}
   149|   149|   149|   149|
   150|   150|   150|   150|	joinedHome := getGroupHome(t, server, "bob-token", aliceGroup.Group.ID)
   151|   151|   151|   151|	if joinedHome.Group.Name != "Breakfast Crew" || joinedHome.Membership.Role != "Player" {
   152|   152|   152|   152|		t.Fatalf("expected Bob to open invited Group home, got %#v", joinedHome)
   153|   153|   153|   153|	}
   154|   154|   154|   154|	stillOwnHome := getGroupHome(t, server, "bob-token", bobExistingGroup.Group.ID)
   155|   155|   155|   155|	if stillOwnHome.Group.Name != "Dinner Weirdos" || stillOwnHome.Membership.Role != "Group Admin" {
   156|   156|   156|   156|		t.Fatalf("expected Bob's existing play history to remain, got %#v", stillOwnHome)
   157|   157|   157|   157|	}
   158|   158|   158|   158|}
   159|   159|   159|   159|
   160|   160|   160|   160|func TestAcceptInviteRejectsAlreadyUsedInvite(t *testing.T) {
   161|   161|   161|   161|	server := newGroupsTestServer()
   162|   162|   162|   162|	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
   163|   163|   163|   163|	invite := createInvite(t, server, "alice-token", aliceGroup.Group.ID)
   164|   164|   164|   164|
   165|   165|   165|   165|	first := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
   166|   166|   166|   166|	if first.Code != http.StatusOK {
   167|   167|   167|   167|		t.Fatalf("expected first accept status 200, got %d: %s", first.Code, first.Body.String())
   168|   168|   168|   168|	}
   169|   169|   169|   169|
   170|   170|   170|   170|	second := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "carol-token", nil)
   171|   171|   171|   171|	if second.Code != http.StatusConflict {
   172|   172|   172|   172|		t.Fatalf("expected already-used Invite status 409, got %d: %s", second.Code, second.Body.String())
   173|   173|   173|   173|	}
   174|   174|   174|   174|}
   175|   175|   175|   175|
   176|   176|   176|   176|func TestAcceptInviteRejectsExistingGroupMemberWithoutUsingInvite(t *testing.T) {
   177|   177|   177|   177|	server := newGroupsTestServer()
   178|   178|   178|   178|	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
   179|   179|   179|   179|	invite := createInvite(t, server, "alice-token", aliceGroup.Group.ID)
   180|   180|   180|   180|
   181|   181|   181|   181|	aliceAccept := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "alice-token", nil)
   182|   182|   182|   182|	if aliceAccept.Code != http.StatusConflict {
   183|   183|   183|   183|		t.Fatalf("expected existing member Invite accept status 409, got %d: %s", aliceAccept.Code, aliceAccept.Body.String())
   184|   184|   184|   184|	}
   185|   185|   185|   185|
   186|   186|   186|   186|	bobAccept := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
   187|   187|   187|   187|	if bobAccept.Code != http.StatusOK {
   188|   188|   188|   188|		t.Fatalf("expected Invite to remain usable for Bob, got %d: %s", bobAccept.Code, bobAccept.Body.String())
   189|   189|   189|   189|	}
   190|   190|   190|   190|}
   191|   191|   191|   191|
   192|   192|   192|   192|func TestAcceptInviteRejectsExpiredInvite(t *testing.T) {
   193|   193|   193|   193|	store := httpapi.NewMemoryStoreWithClock(func() time.Time {
   194|   194|   194|   194|		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
   195|   195|   195|   195|	})
   196|   196|   196|   196|	server := newGroupsTestServerWithStore(store)
   197|   197|   197|   197|	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
   198|   198|   198|   198|	invite := createInvite(t, server, "alice-token", aliceGroup.Group.ID)
   199|   199|   199|   199|
   200|   200|   200|   200|	store.SetClock(func() time.Time {
   201|   201|   201|   201|		return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
   202|   202|   202|   202|	})
   203|   203|   203|   203|	rec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
   204|   204|   204|   204|	if rec.Code != http.StatusGone {
   205|   205|   205|   205|		t.Fatalf("expected expired Invite status 410, got %d: %s", rec.Code, rec.Body.String())
   206|   206|   206|   206|	}
   207|   207|   207|   207|}
   208|   208|   208|   208|
   209|   209|   209|   209|func TestAcceptInviteRejectsInvalidInviteToken(t *testing.T) {
   210|   210|   210|   210|	server := newGroupsTestServer()
   211|   211|   211|   211|
   212|   212|   212|   212|	rec := doJSON(server, http.MethodPost, "/v1/invites/not-a-real-invite/accept", "bob-token", nil)
   213|   213|   213|   213|	if rec.Code != http.StatusNotFound {
   214|   214|   214|   214|		t.Fatalf("expected invalid Invite status 404, got %d: %s", rec.Code, rec.Body.String())
   215|   215|   215|   215|	}
   216|   216|   216|   216|}
   217|   217|   217|   217|
   218|   218|   218|   218|func TestAcceptInviteReturnsStandingsForFinalizedSeason(t *testing.T) {
   219|   219|   219|   219|	store := httpapi.NewMemoryStoreWithClock(func() time.Time {
   220|   220|   220|   220|		return time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
   221|   221|   221|   221|	})
   222|   222|   222|   222|	server := newGroupsTestServerWithStore(store)
   223|   223|   223|   223|	group := createGroup(t, server, "alice-token", "Breakfast Crew")
   224|   224|   224|   224|	startSeasonWithDeadlines(
   225|   225|   225|   225|		t,
   226|   226|   226|   226|		server,
   227|   227|   227|   227|		"alice-token",
   228|   228|   228|   228|		group.Group.ID,
   229|   229|   229|   229|		time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
   230|   230|   230|   230|		time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
   231|   231|   231|   231|	)
   232|   232|   232|   232|	invite := createInvite(t, server, "alice-token", group.Group.ID)
   233|   233|   233|   233|	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
   234|   234|   234|   234|	if acceptRec.Code != http.StatusOK {
   235|   235|   235|   235|		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
   236|   236|   236|   236|	}
   237|   237|   237|   237|	performed := performJump(t, server, "alice-token", group.Group.ID)
   238|   238|   238|   238|	submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 5, 3, 2, http.StatusCreated)
   239|   239|   239|   239|
   240|   240|   240|   240|	store.SetClock(func() time.Time {
   241|   241|   241|   241|		return time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
   242|   242|   242|   242|	})
   243|   243|   243|   243|
   244|   244|   244|   244|	carolAcceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+createInvite(t, server, "alice-token", group.Group.ID).Token+"/accept", "carol-token", nil)
   245|   245|   245|   245|	if carolAcceptRec.Code != http.StatusOK {
   246|   246|   246|   246|		t.Fatalf("expected Carol to join Group after judging, got %d: %s", carolAcceptRec.Code, carolAcceptRec.Body.String())
   247|   247|   247|   247|	}
   248|   248|   248|   248|	var accepted groupHomeBody
   249|   249|   249|   249|	decodeResponse(t, carolAcceptRec, &accepted)
   250|   250|   250|   250|	if len(accepted.Standings) != 1 {
   251|   251|   251|   251|		t.Fatalf("expected standings in Invite accept response, got %#v", accepted.Standings)
   252|   252|   252|   252|	}
   253|   253|   253|   253|	if accepted.Standings[0].Player.ID != group.Membership.PlayerID || accepted.Standings[0].SeasonScore != 14 || accepted.Standings[0].JudgedJumps != 1 {
   254|   254|   254|   254|		t.Fatalf("expected Alice standings in Invite accept response, got %#v", accepted.Standings[0])
   255|   255|   255|   255|	}
   256|   256|   256|   256|}
   257|   257|   257|   257|
   258|   258|   258|   258|func TestCreateInviteRejectsSignedInNonMember(t *testing.T) {
   259|   259|   259|   259|	server := newGroupsTestServer()
   260|   260|   260|   260|	aliceGroup := createGroup(t, server, "alice-token", "Breakfast Crew")
   261|   261|   261|   261|
   262|   262|   262|   262|	rec := doJSON(server, http.MethodPost, "/v1/groups/"+aliceGroup.Group.ID+"/invites", "bob-token", nil)
   263|   263|   263|   263|	if rec.Code != http.StatusForbidden {
   264|   264|   264|   264|		t.Fatalf("expected non-member Invite creation status 403, got %d: %s", rec.Code, rec.Body.String())
   265|   265|   265|   265|	}
   266|   266|   266|   266|}
   267|   267|   267|   267|
   268|   268|   268|   268|func TestGroupMemberCanStartSeasonAndSeeActiveSeasonOnGroupHome(t *testing.T) {
   269|   269|   269|   269|	server := newGroupsTestServer()
   270|   270|   270|   270|	group := createGroup(t, server, "alice-token", "Breakfast Crew")
   271|   271|   271|   271|
   272|   272|   272|   272|	startRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/seasons", "alice-token", map[string]string{
   273|   273|   273|   273|		"submissionDeadline": time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
   274|   274|   274|   274|		"judgingDeadline":    time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339),
   275|   275|   275|   275|	})
   276|   276|   276|   276|	if startRec.Code != http.StatusCreated {
   277|   277|   277|   277|		t.Fatalf("expected status 201, got %d: %s", startRec.Code, startRec.Body.String())
   278|   278|   278|   278|	}
   279|   279|   279|   279|
   280|   280|   280|   280|	var started groupHomeBody
   281|   281|   281|   281|	decodeResponse(t, startRec, &started)
   282|   282|   282|   282|	if started.ActiveSeason == nil {
   283|   283|   283|   283|		t.Fatalf("expected Active Season on start response")
   284|   284|   284|   284|	}
   285|   285|   285|   285|	if started.ActiveSeason.Status != "Active" {
   286|   286|   286|   286|		t.Fatalf("expected Active Season status, got %q", started.ActiveSeason.Status)
   287|   287|   287|   287|	}
   288|   288|   288|   288|	if started.ActiveSeason.CommissionerPlayerID != group.Membership.PlayerID {
   289|   289|   289|   289|		t.Fatalf("expected starting Player to be Season Commissioner, got %#v", started.ActiveSeason)
   290|   290|   290|   290|	}
   291|   291|   291|   291|	if len(started.Standings) != 0 {
   292|   292|   292|   292|		t.Fatalf("expected empty Standings, got %#v", started.Standings)
   293|   293|   293|   293|	}
   294|   294|   294|   294|
   295|   295|   295|   295|	home := getGroupHome(t, server, "alice-token", group.Group.ID)
   296|   296|   296|   296|	if home.ActiveSeason == nil || home.ActiveSeason.ID != started.ActiveSeason.ID {
   297|   297|   297|   297|		t.Fatalf("expected Group home to show backend Active Season, got %#v", home.ActiveSeason)
   298|   298|   298|   298|	}
   299|   299|   299|   299|	if len(home.Standings) != 0 {
   300|   300|   300|   300|		t.Fatalf("expected empty Standings on Group home, got %#v", home.Standings)
   301|   301|   301|   301|	}
   302|   302|   302|   302|}
   303|   303|   303|   303|
   304|   304|   304|   304|func TestSeasonCommissionerCanCloseSubmissionsAndMoveSeasonIntoJudgingGracePeriod(t *testing.T) {
   305|   305|   305|   305|	server := newGroupsTestServer()
   306|   306|   306|   306|	group := createGroup(t, server, "alice-token", "Breakfast Crew")
   307|   307|   307|   307|	season := startSeason(t, server, "alice-token", group.Group.ID)
   308|   308|   308|   308|	idea := createIdea(t, server, "alice-token", group.Group.ID, "Waffle House", "movie theater", "hash browns")
   309|   309|   309|   309|	planned := createPlannedJump(t, server, "alice-token", idea.ID, false)
   310|   310|   310|   310|	authorization := authorizeEvidenceUpload(t, server, "alice-token", planned.ID, "image/jpeg")
   311|   311|   311|   311|
   312|   312|   312|   312|	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "alice-token", nil)
   313|   313|   313|   313|	if closeRec.Code != http.StatusOK {
   314|   314|   314|   314|		t.Fatalf("expected status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
   315|   315|   315|   315|	}
   316|   316|   316|   316|	var closed groupHomeBody
   317|   317|   317|   317|	decodeResponse(t, closeRec, &closed)
   318|   318|   318|   318|	if closed.ActiveSeason == nil || closed.ActiveSeason.ID != season.ActiveSeason.ID || closed.ActiveSeason.Status != "Judging Grace Period" {
   319|   319|   319|   319|		t.Fatalf("expected Season to move into Judging Grace Period, got %#v", closed.ActiveSeason)
   320|   320|   320|   320|	}
   321|   321|   321|   321|
   322|   322|   322|   322|	submitRec := doJSON(server, http.MethodPost, "/v1/jumps/"+planned.ID+"/evidence", "alice-token", map[string]string{
   323|   323|   323|   323|		"uploadAuthorizationId": authorization.ID,
   324|   324|   324|   324|		"caption":               "Too late for competition.",
   325|   325|   325|   325|	})
   326|   326|   326|   326|	if submitRec.Code != http.StatusConflict {
   327|   327|   327|   327|		t.Fatalf("expected closed submission status 409, got %d: %s", submitRec.Code, submitRec.Body.String())
   328|   328|   328|   328|	}
   329|   329|   329|   329|}
   330|   330|   330|   330|
   331|   331|   331|   331|func TestJudgingGracePeriodStillAllowsEligibleJudgmentsOnExistingPerformedJumps(t *testing.T) {
   332|   332|   332|   332|	server := newGroupsTestServer()
   333|   333|   333|   333|	group := createGroup(t, server, "alice-token", "Breakfast Crew")
   334|   334|   334|   334|	season := startSeason(t, server, "alice-token", group.Group.ID)
   335|   335|   335|   335|	invite := createInvite(t, server, "alice-token", group.Group.ID)
   336|   336|   336|   336|	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
   337|   337|   337|   337|	if acceptRec.Code != http.StatusOK {
   338|   338|   338|   338|		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
   339|   339|   339|   339|	}
   340|   340|   340|   340|	performed := performJump(t, server, "alice-token", group.Group.ID)
   341|   341|   341|   341|
   342|   342|   342|   342|	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "alice-token", nil)
   343|   343|   343|   343|	if closeRec.Code != http.StatusOK {
   344|   344|   344|   344|		t.Fatalf("expected status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
   345|   345|   345|   345|	}
   346|   346|   346|   346|
   347|   347|   347|   347|	judgment := submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 5, 3, 2, http.StatusCreated)
   348|   348|   348|   348|	if judgment.JumpID != performed.Jump.ID || judgment.PlayerID == performed.Jump.PlayerID {
   349|   349|   349|   349|		t.Fatalf("expected eligible Judge to score existing Performed Jump during Judging Grace Period, got %#v", judgment)
   350|   350|   350|   350|	}
   351|   351|   351|   351|}
   352|   352|   352|   352|
   353|   353|   353|   353|func TestSeasonCommissionerCanFinalizeSeasonAndLockStandings(t *testing.T) {
   354|   354|   354|   354|	server := newGroupsTestServer()
   355|   355|   355|   355|	group := createGroup(t, server, "alice-token", "Breakfast Crew")
   356|   356|   356|   356|	season := startSeason(t, server, "alice-token", group.Group.ID)
   357|   357|   357|   357|	invite := createInvite(t, server, "alice-token", group.Group.ID)
   358|   358|   358|   358|	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "bob-token", nil)
   359|   359|   359|   359|	if acceptRec.Code != http.StatusOK {
   360|   360|   360|   360|		t.Fatalf("expected Bob to join Group before judging, got %d: %s", acceptRec.Code, acceptRec.Body.String())
   361|   361|   361|   361|	}
   362|   362|   362|   362|	performed := performJump(t, server, "alice-token", group.Group.ID)
   363|   363|   363|   363|	submitJudgment(t, server, "bob-token", performed.Jump.ID, 4, 5, 3, 2, http.StatusCreated)
   364|   364|   364|   364|
   365|   365|   365|   365|	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "alice-token", nil)
   366|   366|   366|   366|	if closeRec.Code != http.StatusOK {
   367|   367|   367|   367|		t.Fatalf("expected status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
   368|   368|   368|   368|	}
   369|   369|   369|   369|
   370|   370|   370|   370|	finalizeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/finalize", "alice-token", nil)
   371|   371|   371|   371|	if finalizeRec.Code != http.StatusOK {
   372|   372|   372|   372|		t.Fatalf("expected status 200, got %d: %s", finalizeRec.Code, finalizeRec.Body.String())
   373|   373|   373|   373|	}
   374|   374|   374|   374|	var finalized groupHomeBody
   375|   375|   375|   375|	decodeResponse(t, finalizeRec, &finalized)
   376|   376|   376|   376|	if finalized.ActiveSeason != nil {
   377|   377|   377|   377|		t.Fatalf("expected Finalized Season to no longer be open, got %#v", finalized.ActiveSeason)
   378|   378|   378|   378|	}
   379|   379|   379|   379|	if len(finalized.Standings) != 1 || finalized.Standings[0].SeasonScore != 14 || finalized.Standings[0].JudgedJumps != 1 {
   380|   380|   380|   380|		t.Fatalf("expected locked Standings with Alice on 14 from one Judged Jump, got %#v", finalized.Standings)
   381|   381|   381|   381|	}
   382|   382|   382|   382|
   383|   383|   383|   383|	editRec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
   384|   384|   384|   384|		"commitment":    10,
   385|   385|   385|   385|		"transgression": 10,
   386|   386|   386|   386|		"creativity":    10,
   387|   387|   387|   387|		"presentation": 10,
   388|   388|   388|   388|	})
   389|   389|   389|   389|	if editRec.Code != http.StatusConflict {
   390|   390|   390|   390|		t.Fatalf("expected Finalized Season to lock remaining Judging Windows, got %d: %s", editRec.Code, editRec.Body.String())
   391|   391|   391|   391|	}
   392|   392|   392|   392|}
   393|   393|   393|   393|
   394|   394|   394|   394|func TestGroupAdminEmergencySeasonOverridesAppearInSeasonHistory(t *testing.T) {
   395|   395|   395|   395|	server := newGroupsTestServer()
   396|   396|   396|   396|	group := createGroup(t, server, "bob-token", "Breakfast Crew")
   397|   397|   397|   397|	invite := createInvite(t, server, "bob-token", group.Group.ID)
   398|   398|   398|   398|	acceptRec := doJSON(server, http.MethodPost, "/v1/invites/"+invite.Token+"/accept", "alice-token", nil)
   399|   399|   399|   399|	if acceptRec.Code != http.StatusOK {
   400|   400|   400|   400|		t.Fatalf("expected Alice to join Group before starting Season, got %d: %s", acceptRec.Code, acceptRec.Body.String())
   401|   401|   401|   401|	}
   402|   402|   402|   402|	season := startSeason(t, server, "alice-token", group.Group.ID)
   403|   403|   403|   403|	performed := performJump(t, server, "alice-token", group.Group.ID)
   404|   404|   404|   404|
   405|   405|   405|   405|	closeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/close-submissions", "bob-token", nil)
   406|   406|   406|   406|	if closeRec.Code != http.StatusOK {
   407|   407|   407|   407|		t.Fatalf("expected Group Admin override close status 200, got %d: %s", closeRec.Code, closeRec.Body.String())
   408|   408|   408|   408|	}
   409|   409|   409|   409|
   410|   410|   410|   410|	finalizeRec := doJSON(server, http.MethodPost, "/v1/seasons/"+season.ActiveSeason.ID+"/finalize", "bob-token", nil)
   411|   411|   411|   411|	if finalizeRec.Code != http.StatusOK {
   412|   412|   412|   412|		t.Fatalf("expected Group Admin override finalize status 200, got %d: %s", finalizeRec.Code, finalizeRec.Body.String())
   413|   413|   413|   413|	}
   414|   414|   414|   414|
   415|   415|   415|   415|	history := getSeasonHistory(t, server, "alice-token", season.ActiveSeason.ID)
   416|   416|   416|   416|	if len(history.Entries) != 2 {
   417|   417|   417|   417|		t.Fatalf("expected two visible override actions in Season history, got %#v", history.Entries)
   418|   418|   418|   418|	}
   419|   419|   419|   419|	if history.Entries[0].Action != "Submissions Closed" || history.Entries[0].ActorPlayerID != group.Membership.PlayerID || !history.Entries[0].Override || history.Entries[0].ActorRole != "Group Admin" {
   420|   420|   420|   420|		t.Fatalf("expected visible Group Admin close override entry, got %#v", history.Entries[0])
   421|   421|   421|   421|	}
   422|   422|   422|   422|	if history.Entries[1].Action != "Season Finalized" || history.Entries[1].ActorPlayerID != group.Membership.PlayerID || !history.Entries[1].Override || history.Entries[1].ActorRole != "Group Admin" {
   423|   423|   423|   423|		t.Fatalf("expected visible Group Admin finalize override entry, got %#v", history.Entries[1])
   424|   424|   424|   424|	}
   425|   425|   425|   425|	if history.Entries[1].ToStatus != "Finalized" {
   426|   426|   426|   426|		t.Fatalf("expected finalized status in Season history, got %#v", history.Entries[1])
   427|   427|   427|   427|	}
   428|   428|   428|   428|
   429|   429|   429|   429|	editRec := doJSON(server, http.MethodPost, "/v1/jumps/"+performed.Jump.ID+"/judgment", "bob-token", map[string]int{
   430|   430|   430|   430|		"commitment":    1,
   431|   431|   431|   431|		"transgression": 1,
   432|   432|   432|   432|		"creativity":    1,
   433|   433|   433|   433|		"presentation": 1,
   434|   434|   434|   434|	})
   435|   435|   435|   435|	if editRec.Code != http.StatusConflict {
   436|   436|   436|   436|		t.Fatalf("expected Group Admin finalization override to cap remaining Judging Windows, got %d: %s", editRec.Code, editRec.Body.String())
   437|   437|   437|   437|	}
   438|   438|   438|   438|}
   439|   439|   439|   439|
   440|   440|   440|   440|func TestGroupMemberCreatesIdeaThenSeasonLinkedPlannedJumpDuringActiveSeason(t *testing.T) {
   441|   441|   441|   441|	server := newGroupsTestServer()
   442|   442|   442|   442|	group := createGroup(t, server, "alice-token", "Breakfast Crew")
   443|   443|   443|   443|	season := startSeason(t, server, "alice-token", group.Group.ID)
   444|   444|   444|   444|
   445|   445|   445|   445|	ideaRec := doJSON(server, http.MethodPost, "/v1/groups/"+group.Group.ID+"/ideas", "alice-token", map[string]string{
   446|   446|   446|   446|		"source":      "Taco Bell",
   447|   447|   447|   447|		"destination": "Olive Garden parking lot",
   448|   448|   448|   448|		"food":        "Crunchwrap",
   449|   449|   449|   449|	})
   450|   450|   450|   450|	if ideaRec.Code != http.StatusCreated {
   451|   451|   451|   451|		t.Fatalf("expected Idea status 201, got %d: %s", ideaRec.Code, ideaRec.Body.String())
   452|   452|   452|   452|	}
   453|   453|   453|   453|	var idea jumpBody
   454|   454|   454|   454|	decodeResponse(t, ideaRec, &idea)
   455|   455|   455|   455|	if idea.ID == "" || idea.GroupID != group.Group.ID || idea.PlayerID != group.Membership.PlayerID {
   456|   456|   456|   456|		t.Fatalf("expected Group-scoped Idea for Alice, got %#v", idea)
   457|   457|   457|   457|	}
   458|   458|   458|   458|	if idea.Status != "Idea" || idea.SeasonID != nil || !idea.OffSeason {
   459|   459|   459|   459|		t.Fatalf("expected Off-Season Idea before planning, got %#v", idea)
   460|   460|   460|   460|	}
   461|   461|   461|   461|
   462|   462|   462|   462|	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-jump", "alice-token", nil)
   463|   463|   463|   463|	if planRec.Code != http.StatusCreated {
   464|   464|   464|   464|		t.Fatalf("expected Planned Jump status 201, got %d: %s", planRec.Code, planRec.Body.String())
   465|   465|   465|   465|	}
   466|   466|   466|   466|	var planned jumpBody
   467|   467|   467|   467|	decodeResponse(t, planRec, &planned)
   468|   468|   468|   468|	if planned.ID != idea.ID || planned.Status != "Planned Jump" {
   469|   469|   469|   469|		t.Fatalf("expected same Jump to become a Planned Jump, got %#v", planned)
   470|   470|   470|   470|	}
   471|   471|   471|   471|	if planned.GroupID != group.Group.ID {
   472|   472|   472|   472|		t.Fatalf("expected Planned Jump to belong to exactly one Group %q, got %q", group.Group.ID, planned.GroupID)
   473|   473|   473|   473|	}
   474|   474|   474|   474|	if planned.SeasonID == nil || *planned.SeasonID != season.ActiveSeason.ID || planned.OffSeason {
   475|   475|   475|   475|		t.Fatalf("expected Planned Jump to be Season-linked by default, got %#v", planned)
   476|   476|   476|   476|	}
   477|   477|   477|   477|	if planned.Source != "Taco Bell" || planned.Destination != "Olive Garden parking lot" || planned.Food != "Crunchwrap" {
   478|   478|   478|   478|		t.Fatalf("expected Source, Destination, and Food from Idea, got %#v", planned)
   479|   479|   479|   479|	}
   480|   480|   480|   480|}
   481|   481|   481|   481|
   482|   482|   482|   482|func TestPlannedJumpCanBeExplicitlyOffSeasonDuringActiveSeason(t *testing.T) {
   483|   483|   483|   483|	server := newGroupsTestServer()
   484|   484|   484|   484|	group := createGroup(t, server, "alice-token", "Breakfast Crew")
   485|   485|   485|   485|	startSeason(t, server, "alice-token", group.Group.ID)
   486|   486|   486|   486|	idea := createIdea(t, server, "alice-token", group.Group.ID, "Waffle House", "movie theater", "hash browns")
   487|   487|   487|   487|
   488|   488|   488|   488|	planRec := doJSON(server, http.MethodPost, "/v1/ideas/"+idea.ID+"/planned-jump", "alice-token", map[string]bool{"offSeason": true})
   489|   489|   489|   489|	if planRec.Code != http.StatusCreated {
   490|   490|   490|   490|		t.Fatalf("expected Planned Jump status 201, got %d: %s", planRec.Code, planRec.Body.String())
   491|   491|   491|   491|	}
   492|   492|   492|   492|	var planned jumpBody
   493|   493|   493|   493|	decodeResponse(t, planRec, &planned)
   494|   494|   494|   494|	if planned.SeasonID != nil || !planned.OffSeason {
   495|   495|   495|   495|		t.Fatalf("expected explicit Off-Season Jump, got %#v", planned)
   496|   496|   496|   496|	}
   497|   497|   497|   497|}
   498|   498|   498|   498|
   499|   499|   499|   499|func TestPlannedJumpIsOffSeasonWhenNoSeasonIsActive(t *testing.T) {
   500|   500|   500|   500|	server := newGroupsTestServer()
   501|