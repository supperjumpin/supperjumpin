     1|     1|     1|     1|     1|package httpapi
     2|     2|     2|     2|     2|
     3|     3|     3|     3|     3|import (
     4|     4|     4|     4|     4|	"context"
     5|     5|     5|     5|     5|	"crypto/rand"
     6|     6|     6|     6|     6|	"crypto/sha256"
     7|     7|     7|     7|     7|	"encoding/hex"
     8|     8|     8|     8|     8|	"errors"
     9|     9|     9|     9|     9|	"fmt"
    10|    10|    10|    10|    10|	"sort"
    11|    11|    11|    11|    11|	"strconv"
    12|    12|    12|    12|    12|	"strings"
    13|    13|    13|    13|    13|	"sync"
    14|    14|    14|    14|    14|	"time"
    15|    15|    15|    15|    15|)
    16|    16|    16|    16|    16|
    17|    17|    17|    17|    17|var ErrSeasonAlreadyOpen = errors.New("Group already has an active or closing Season")
    18|    18|    18|    18|    18|
    19|    19|    19|    19|    19|var ErrSeasonNotFound = errors.New("Season not found")
    20|    20|    20|    20|    20|
    21|    21|    21|    21|    21|var ErrStuntNotFound = errors.New("Stunt not found")
    22|    22|    22|    22|    22|
    23|    23|    23|    23|    23|var ErrEvidenceUploadAuthorizationNotFound = errors.New("Evidence upload authorization not found")
    24|    24|    24|    24|    24|
    25|    25|    25|    25|    25|var ErrJudgingWindowClosed = errors.New("Judging Window closed")
    26|    26|    26|    26|    26|
    27|    27|    27|    27|    27|var ErrSubmissionWindowClosed = errors.New("Submission Window closed")
    28|    28|    28|    28|    28|
    29|    29|    29|    29|    29|var ErrInvalidJudgmentScore = errors.New("Judgment scores must be between 0 and 10")
    30|    30|    30|    30|    30|
    31|    31|    31|    31|    31|var ErrInvalidDisputeConcern = errors.New("Dispute concern must be House Rules, Credibility, Source, Destination, Food, duplicate, or other")
    32|    32|    32|    32|    32|
    33|    33|    33|    33|    33|var ErrDisputeNotFound = errors.New("Dispute not found")
    34|    34|    34|    34|    34|
    35|    35|    35|    35|    35|var ErrInvalidDisputeResolution = errors.New("Dispute resolution must be No Action, Disqualified Jump, or Removed Jump")
    36|    36|    36|    36|    36|
    37|    37|    37|    37|    37|type Account struct {
    38|    38|    38|    38|    38|	ID    string `json:"id"`
    39|    39|    39|    39|    39|	Email string `json:"email"`
    40|    40|    40|    40|    40|}
    41|    41|    41|    41|    41|
    42|    42|    42|    42|    42|type Player struct {
    43|    43|    43|    43|    43|	ID          string `json:"id"`
    44|    44|    44|    44|    44|	DisplayName string `json:"displayName"`
    45|    45|    45|    45|    45|}
    46|    46|    46|    46|    46|
    47|    47|    47|    47|    47|type MeResponse struct {
    48|    48|    48|    48|    48|	Account Account `json:"account"`
    49|    49|    49|    49|    49|	Player  Player  `json:"player"`
    50|    50|    50|    50|    50|}
    51|    51|    51|    51|    51|
    52|    52|    52|    52|    52|type Group struct {
    53|    53|    53|    53|    53|	ID   string `json:"id"`
    54|    54|    54|    54|    54|	Name string `json:"name"`
    55|    55|    55|    55|    55|}
    56|    56|    56|    56|    56|
    57|    57|    57|    57|    57|type GroupMembership struct {
    58|    58|    58|    58|    58|	GroupID  string `json:"groupId"`
    59|    59|    59|    59|    59|	PlayerID string `json:"playerId"`
    60|    60|    60|    60|    60|	Role     string `json:"role"`
    61|    61|    61|    61|    61|}
    62|    62|    62|    62|    62|
    63|    63|    63|    63|    63|type Season struct {
    64|    64|    64|    64|    64|	ID                   string    `json:"id"`
    65|    65|    65|    65|    65|	GroupID              string    `json:"groupId"`
    66|    66|    66|    66|    66|	CommissionerPlayerID string    `json:"commissionerPlayerId"`
    67|    67|    67|    67|    67|	Status               string    `json:"status"`
    68|    68|    68|    68|    68|	SubmissionDeadline   time.Time `json:"submissionDeadline"`
    69|    69|    69|    69|    69|	JudgingDeadline      time.Time `json:"judgingDeadline"`
    70|    70|    70|    70|    70|}
    71|    71|    71|    71|    71|
    72|    72|    72|    72|    72|type SeasonHistoryEntry struct {
    73|    73|    73|    73|    73|	ID            string    `json:"id"`
    74|    74|    74|    74|    74|	SeasonID      string    `json:"seasonId"`
    75|    75|    75|    75|    75|	Action        string    `json:"action"`
    76|    76|    76|    76|    76|	ActorPlayerID string    `json:"actorPlayerId"`
    77|    77|    77|    77|    77|	ActorRole     string    `json:"actorRole"`
    78|    78|    78|    78|    78|	Override      bool      `json:"override"`
    79|    79|    79|    79|    79|	FromStatus    string    `json:"fromStatus"`
    80|    80|    80|    80|    80|	ToStatus      string    `json:"toStatus"`
    81|    81|    81|    81|    81|	CreatedAt     time.Time `json:"createdAt"`
    82|    82|    82|    82|    82|}
    83|    83|    83|    83|    83|
    84|    84|    84|    84|    84|type SeasonHistoryResponse struct {
    85|    85|    85|    85|    85|	Entries []SeasonHistoryEntry `json:"entries"`
    86|    86|    86|    86|    86|}
    87|    87|    87|    87|    87|
    88|    88|    88|    88|    88|type GroupHomeResponse struct {
    89|    89|    89|    89|    89|	Group        Group                `json:"group"`
    90|    90|    90|    90|    90|	Membership   GroupMembership      `json:"membership"`
    91|    91|    91|    91|    91|	ActiveSeason *Season              `json:"activeSeason"`
    92|    92|    92|    92|    92|	RecentStunts []PerformedStuntView `json:"recentStunts"`
    93|    93|    93|    93|    93|	Standings    []StandingEntry      `json:"standings"`
    94|    94|    94|    94|    94|}
    95|    95|    95|    95|    95|
    96|    96|    96|    96|    96|type StandingEntry struct {
    97|    97|    97|    97|    97|	Player       Player `json:"player"`
    98|    98|    98|    98|    98|	SeasonScore  int    `json:"seasonScore"`
    99|    99|    99|    99|    99|	JudgedStunts int    `json:"judgedStunts"`
   100|   100|   100|   100|   100|}
   101|   101|   101|   101|   101|
   102|   102|   102|   102|   102|type PerformedStuntView struct {
   103|   103|   103|   103|   103|	Stunt     Stunt     `json:"stunt"`
   104|   104|   104|   104|   104|	Performer Player    `json:"performer"`
   105|   105|   105|   105|   105|	Evidence  Evidence  `json:"evidence"`
   106|   106|   106|   106|   106|	Disputes  []Dispute `json:"disputes"`
   107|   107|   107|   107|   107|}
   108|   108|   108|   108|   108|
   109|   109|   109|   109|   109|type Stunt struct {
   110|   110|   110|   110|   110|	ID          string  `json:"id"`
   111|   111|   111|   111|   111|	GroupID     string  `json:"groupId"`
   112|   112|   112|   112|   112|	PlayerID    string  `json:"playerId"`
   113|   113|   113|   113|   113|	SeasonID    *string `json:"seasonId"`
   114|   114|   114|   114|   114|	Status      string  `json:"status"`
   115|   115|   115|   115|   115|	Source      string  `json:"source"`
   116|   116|   116|   116|   116|	Destination string  `json:"destination"`
   117|   117|   117|   117|   117|	Food        string  `json:"food"`
   118|   118|   118|   118|   118|	OffSeason   bool    `json:"offSeason"`
   119|   119|   119|   119|   119|	FinalScore  *int    `json:"finalScore"`
   120|   120|   120|   120|   120|}
   121|   121|   121|   121|   121|
   122|   122|   122|   122|   122|type EvidenceUploadAuthorization struct {
   123|   123|   123|   123|   123|	ID             string            `json:"id"`
   124|   124|   124|   124|   124|	StuntID        string            `json:"stuntId"`
   125|   125|   125|   125|   125|	UploadURL      string            `json:"uploadUrl"`
   126|   126|   126|   126|   126|	UploadMethod   string            `json:"uploadMethod"`
   127|   127|   127|   127|   127|	UploadHeaders  map[string]string `json:"uploadHeaders"`
   128|   128|   128|   128|   128|	MediaObjectKey string            `json:"mediaObjectKey"`
   129|   129|   129|   129|   129|	ExpiresAt      time.Time         `json:"expiresAt"`
   130|   130|   130|   130|   130|}
   131|   131|   131|   131|   131|
   132|   132|   132|   132|   132|type Evidence struct {
   133|   133|   133|   133|   133|	ID             string    `json:"id"`
   134|   134|   134|   134|   134|	StuntID        string    `json:"stuntId"`
   135|   135|   135|   135|   135|	Caption        string    `json:"caption"`
   136|   136|   136|   136|   136|	MediaObjectKey string    `json:"mediaObjectKey"`
   137|   137|   137|   137|   137|	CreatedAt      time.Time `json:"createdAt"`
   138|   138|   138|   138|   138|}
   139|   139|   139|   139|   139|
   140|   140|   140|   140|   140|type EvidenceSubmission struct {
   141|   141|   141|   141|   141|	Stunt    Stunt    `json:"stunt"`
   142|   142|   142|   142|   142|	Evidence Evidence `json:"evidence"`
   143|   143|   143|   143|   143|}
   144|   144|   144|   144|   144|
   145|   145|   145|   145|   145|type Judgment struct {
   146|   146|   146|   146|   146|	ID            string `json:"id"`
   147|   147|   147|   147|   147|	StuntID       string `json:"stuntId"`
   148|   148|   148|   148|   148|	PlayerID      string `json:"playerId"`
   149|   149|   149|   149|   149|	Commitment    int    `json:"commitment"`
   150|   150|   150|   150|   150|	Transgression int    `json:"transgression"`
   151|   151|   151|   151|   151|	Creativity    int    `json:"creativity"`
   152|   152|   152|   152|   152|	Documentation int    `json:"documentation"`
   153|   153|   153|   153|   153|}
   154|   154|   154|   154|   154|
   155|   155|   155|   155|   155|type Dispute struct {
   156|   156|   156|   156|   156|	ID                 string  `json:"id"`
   157|   157|   157|   157|   157|	StuntID            string  `json:"stuntId"`
   158|   158|   158|   158|   158|	RaisedByPlayerID   string  `json:"raisedByPlayerId"`
   159|   159|   159|   159|   159|	Concern            string  `json:"concern"`
   160|   160|   160|   160|   160|	Details            string  `json:"details"`
   161|   161|   161|   161|   161|	Status             string  `json:"status"`
   162|   162|   162|   162|   162|	Resolution         *string `json:"resolution"`
   163|   163|   163|   163|   163|	ResolutionReason   *string `json:"resolutionReason"`
   164|   164|   164|   164|   164|	ResolvedByPlayerID *string `json:"resolvedByPlayerId"`
   165|   165|   165|   165|   165|	OverrideResolution *string `json:"overrideResolution"`
   166|   166|   166|   166|   166|	OverrideReason     *string `json:"overrideReason"`
   167|   167|   167|   167|   167|	OverrideByPlayerID *string `json:"overrideByPlayerId"`
   168|   168|   168|   168|   168|}
   169|   169|   169|   169|   169|
   170|   170|   170|   170|   170|type DisputeResolution struct {
   171|   171|   171|   171|   171|	Stunt   Stunt   `json:"stunt"`
   172|   172|   172|   172|   172|	Dispute Dispute `json:"dispute"`
   173|   173|   173|   173|   173|}
   174|   174|   174|   174|   174|
   175|   175|   175|   175|   175|type GroupMembershipSummary struct {
   176|   176|   176|   176|   176|	Group      Group           `json:"group"`
   177|   177|   177|   177|   177|	Membership GroupMembership `json:"membership"`
   178|   178|   178|   178|   178|}
   179|   179|   179|   179|   179|
   180|   180|   180|   180|   180|type ListGroupsResponse struct {
   181|   181|   181|   181|   181|	Memberships []GroupMembershipSummary `json:"memberships"`
   182|   182|   182|   182|   182|}
   183|   183|   183|   183|   183|
   184|   184|   184|   184|   184|type Invite struct {
   185|   185|   185|   185|   185|	ID        string    `json:"id"`
   186|   186|   186|   186|   186|	GroupID   string    `json:"groupId"`
   187|   187|   187|   187|   187|	Token     string    `json:"token"`
   188|   188|   188|   188|   188|	CreatedBy string    `json:"createdBy"`
   189|   189|   189|   189|   189|	ExpiresAt time.Time `json:"expiresAt"`
   190|   190|   190|   190|   190|}
   191|   191|   191|   191|   191|
   192|   192|   192|   192|   192|type InviteAcceptStatus string
   193|   193|   193|   193|   193|
   194|   194|   194|   194|   194|const (
   195|   195|   195|   195|   195|	InviteAccepted InviteAcceptStatus = "accepted"
   196|   196|   196|   196|   196|	InviteInvalid  InviteAcceptStatus = "invalid"
   197|   197|   197|   197|   197|	InviteUsed     InviteAcceptStatus = "used"
   198|   198|   198|   198|   198|	InviteExpired  InviteAcceptStatus = "expired"
   199|   199|   199|   199|   199|	InviteMember   InviteAcceptStatus = "member"
   200|   200|   200|   200|   200|)
   201|   201|   201|   201|   201|
   202|   202|   202|   202|   202|type Store interface {
   203|   203|   203|   203|   203|	BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error)
   204|   204|   204|   204|   204|	CreateGroup(ctx context.Context, player Player, name string) (GroupHomeResponse, error)
   205|   205|   205|   205|   205|	GroupHome(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error)
   206|   206|   206|   206|   206|	ListGroups(ctx context.Context, player Player) (ListGroupsResponse, error)
   207|   207|   207|   207|   207|	CreateInvite(ctx context.Context, player Player, groupID string) (Invite, bool, error)
   208|   208|   208|   208|   208|	AcceptInvite(ctx context.Context, player Player, token string) (GroupHomeResponse, InviteAcceptStatus, error)
   209|   209|   209|   209|   209|	StartSeason(ctx context.Context, player Player, groupID string, submissionDeadline time.Time, judgingDeadline time.Time) (GroupHomeResponse, bool, error)
   210|   210|   210|   210|   210|	CloseSeasonSubmissions(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error)
   211|   211|   211|   211|   211|	FinalizeSeason(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error)
   212|   212|   212|   212|   212|	SeasonHistory(ctx context.Context, player Player, seasonID string) (SeasonHistoryResponse, bool, error)
   213|   213|   213|   213|   213|	CreateIdea(ctx context.Context, player Player, groupID string, source string, destination string, food string) (Stunt, bool, error)
   214|   214|   214|   214|   214|	CreatePlannedStunt(ctx context.Context, player Player, ideaID string, offSeason bool) (Stunt, bool, error)
   215|   215|   215|   215|   215|	AuthorizeEvidenceUpload(ctx context.Context, player Player, stuntID string, contentType string) (EvidenceUploadAuthorization, bool, error)
   216|   216|   216|   216|   216|	SubmitEvidence(ctx context.Context, player Player, stuntID string, uploadAuthorizationID string, caption string) (EvidenceSubmission, bool, error)
   217|   217|   217|   217|   217|	SubmitJudgment(ctx context.Context, player Player, stuntID string, commitment int, transgression int, creativity int, documentation int) (Judgment, bool, bool, error)
   218|   218|   218|   218|   218|	CreateDispute(ctx context.Context, player Player, stuntID string, concern string, details string) (Dispute, bool, error)
   219|   219|   219|   219|   219|	ResolveDispute(ctx context.Context, player Player, disputeID string, resolution string, resolutionReason string) (DisputeResolution, bool, error)
   220|   220|   220|   220|   220|}
   221|   221|   221|   221|   221|
   222|   222|   222|   222|   222|type MemoryStore struct {
   223|   223|   223|   223|   223|	mu                sync.Mutex
   224|   224|   224|   224|   224|	accounts          map[string]MeResponse
   225|   225|   225|   225|   225|	players           map[string]Player
   226|   226|   226|   226|   226|	groups            map[string]Group
   227|   227|   227|   227|   227|	memberships       map[string]map[string]GroupMembership
   228|   228|   228|   228|   228|	invites           map[string]memoryInvite
   229|   229|   229|   229|   229|	seasons           map[string]Season
   230|   230|   230|   230|   230|	seasonEvents      map[string][]SeasonHistoryEntry
   231|   231|   231|   231|   231|	seasonOrder       map[string]int
   232|   232|   232|   232|   232|	stunts            map[string]Stunt
   233|   233|   233|   233|   233|	uploads           map[string]EvidenceUploadAuthorization
   234|   234|   234|   234|   234|	evidences         map[string]Evidence
   235|   235|   235|   235|   235|	judgments         map[string]Judgment
   236|   236|   236|   236|   236|	disputes          map[string]Dispute
   237|   237|   237|   237|   237|	now               func() time.Time
   238|   238|   238|   238|   238|	groupNumber       int
   239|   239|   239|   239|   239|	inviteNumber      int
   240|   240|   240|   240|   240|	seasonNumber      int
   241|   241|   241|   241|   241|	stuntNumber       int
   242|   242|   242|   242|   242|	uploadNumber      int
   243|   243|   243|   243|   243|	seasonEventNumber int
   244|   244|   244|   244|   244|}
   245|   245|   245|   245|   245|
   246|   246|   246|   246|   246|type memoryInvite struct {
   247|   247|   247|   247|   247|	Invite
   248|   248|   248|   248|   248|	UsedBy string
   249|   249|   249|   249|   249|}
   250|   250|   250|   250|   250|
   251|   251|   251|   251|   251|func NewMemoryStore() *MemoryStore {
   252|   252|   252|   252|   252|	return NewMemoryStoreWithClock(time.Now)
   253|   253|   253|   253|   253|}
   254|   254|   254|   254|   254|
   255|   255|   255|   255|   255|func NewMemoryStoreWithClock(now func() time.Time) *MemoryStore {
   256|   256|   256|   256|   256|	return &MemoryStore{
   257|   257|   257|   257|   257|		accounts:     map[string]MeResponse{},
   258|   258|   258|   258|   258|		players:      map[string]Player{},
   259|   259|   259|   259|   259|		groups:       map[string]Group{},
   260|   260|   260|   260|   260|		memberships:  map[string]map[string]GroupMembership{},
   261|   261|   261|   261|   261|		invites:      map[string]memoryInvite{},
   262|   262|   262|   262|   262|		seasons:      map[string]Season{},
   263|   263|   263|   263|   263|		seasonEvents: map[string][]SeasonHistoryEntry{},
   264|   264|   264|   264|   264|		seasonOrder:  map[string]int{},
   265|   265|   265|   265|   265|		stunts:       map[string]Stunt{},
   266|   266|   266|   266|   266|		uploads:      map[string]EvidenceUploadAuthorization{},
   267|   267|   267|   267|   267|		evidences:    map[string]Evidence{},
   268|   268|   268|   268|   268|		judgments:    map[string]Judgment{},
   269|   269|   269|   269|   269|		disputes:     map[string]Dispute{},
   270|   270|   270|   270|   270|		now:          now,
   271|   271|   271|   271|   271|	}
   272|   272|   272|   272|   272|}
   273|   273|   273|   273|   273|
   274|   274|   274|   274|   274|func (s *MemoryStore) SetClock(now func() time.Time) {
   275|   275|   275|   275|   275|	s.mu.Lock()
   276|   276|   276|   276|   276|	defer s.mu.Unlock()
   277|   277|   277|   277|   277|	s.now = now
   278|   278|   278|   278|   278|}
   279|   279|   279|   279|   279|
   280|   280|   280|   280|   280|func (s *MemoryStore) SetSeasonStatus(seasonID string, status string) {
   281|   281|   281|   281|   281|	s.mu.Lock()
   282|   282|   282|   282|   282|	defer s.mu.Unlock()
   283|   283|   283|   283|   283|	season := s.seasons[seasonID]
   284|   284|   284|   284|   284|	season.Status = status
   285|   285|   285|   285|   285|	s.seasons[seasonID] = season
   286|   286|   286|   286|   286|}
   287|   287|   287|   287|   287|
   288|   288|   288|   288|   288|func (s *MemoryStore) BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error) {
   289|   289|   289|   289|   289|	s.mu.Lock()
   290|   290|   290|   290|   290|	defer s.mu.Unlock()
   291|   291|   291|   291|   291|
   292|   292|   292|   292|   292|	key := identity.Provider + ":" + identity.Subject
   293|   293|   293|   293|   293|	if profile, ok := s.accounts[key]; ok {
   294|   294|   294|   294|   294|		return profile, nil
   295|   295|   295|   295|   295|	}
   296|   296|   296|   296|   296|
   297|   297|   297|   297|   297|	accountID := stableID("account", key)
   298|   298|   298|   298|   298|	profile := MeResponse{
   299|   299|   299|   299|   299|		Account: Account{ID: accountID, Email: identity.Email},
   300|   300|   300|   300|   300|		Player:  Player{ID: stableID("player", accountID), DisplayName: displayName(identity.Email)},
   301|   301|   301|   301|   301|	}
   302|   302|   302|   302|   302|	s.accounts[key] = profile
   303|   303|   303|   303|   303|	s.players[profile.Player.ID] = profile.Player
   304|   304|   304|   304|   304|	return profile, nil
   305|   305|   305|   305|   305|}
   306|   306|   306|   306|   306|
   307|   307|   307|   307|   307|func (s *MemoryStore) CreateGroup(ctx context.Context, player Player, name string) (GroupHomeResponse, error) {
   308|   308|   308|   308|   308|	s.mu.Lock()
   309|   309|   309|   309|   309|	defer s.mu.Unlock()
   310|   310|   310|   310|   310|
   311|   311|   311|   311|   311|	s.groupNumber++
   312|   312|   312|   312|   312|	group := Group{ID: stableID("group", player.ID+":"+name+":"+strconv.Itoa(s.groupNumber)), Name: name}
   313|   313|   313|   313|   313|	membership := GroupMembership{GroupID: group.ID, PlayerID: player.ID, Role: "Group Admin"}
   314|   314|   314|   314|   314|	s.groups[group.ID] = group
   315|   315|   315|   315|   315|	if s.memberships[group.ID] == nil {
   316|   316|   316|   316|   316|		s.memberships[group.ID] = map[string]GroupMembership{}
   317|   317|   317|   317|   317|	}
   318|   318|   318|   318|   318|	s.memberships[group.ID][player.ID] = membership
   319|   319|   319|   319|   319|	return groupHome(group, membership, nil, s.recentPerformedStuntsForGroup(group.ID), s.standingsForGroup(group.ID)), nil
   320|   320|   320|   320|   320|}
   321|   321|   321|   321|   321|
   322|   322|   322|   322|   322|func (s *MemoryStore) GroupHome(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error) {
   323|   323|   323|   323|   323|	s.mu.Lock()
   324|   324|   324|   324|   324|	defer s.mu.Unlock()
   325|   325|   325|   325|   325|
   326|   326|   326|   326|   326|	group, ok := s.groups[groupID]
   327|   327|   327|   327|   327|	if !ok {
   328|   328|   328|   328|   328|		return GroupHomeResponse{}, false, nil
   329|   329|   329|   329|   329|	}
   330|   330|   330|   330|   330|	membership, ok := s.memberships[groupID][player.ID]
   331|   331|   331|   331|   331|	if !ok {
   332|   332|   332|   332|   332|		return GroupHomeResponse{}, false, nil
   333|   333|   333|   333|   333|	}
   334|   334|   334|   334|   334|	s.ensureSeasonStatusesForGroup(groupID)
   335|   335|   335|   335|   335|	season := s.currentSeasonForGroup(groupID)
   336|   336|   336|   336|   336|	return groupHome(group, membership, season, s.recentPerformedStuntsForGroup(groupID), s.standingsForGroup(groupID)), true, nil
   337|   337|   337|   337|   337|}
   338|   338|   338|   338|   338|
   339|   339|   339|   339|   339|func (s *MemoryStore) ListGroups(ctx context.Context, player Player) (ListGroupsResponse, error) {
   340|   340|   340|   340|   340|	s.mu.Lock()
   341|   341|   341|   341|   341|	defer s.mu.Unlock()
   342|   342|   342|   342|   342|
   343|   343|   343|   343|   343|	memberships := []GroupMembershipSummary{}
   344|   344|   344|   344|   344|	for groupID, groupMemberships := range s.memberships {
   345|   345|   345|   345|   345|		membership, ok := groupMemberships[player.ID]
   346|   346|   346|   346|   346|		if !ok {
   347|   347|   347|   347|   347|			continue
   348|   348|   348|   348|   348|		}
   349|   349|   349|   349|   349|		memberships = append(memberships, GroupMembershipSummary{Group: s.groups[groupID], Membership: membership})
   350|   350|   350|   350|   350|	}
   351|   351|   351|   351|   351|	sort.Slice(memberships, func(i, j int) bool {
   352|   352|   352|   352|   352|		return memberships[i].Group.Name < memberships[j].Group.Name
   353|   353|   353|   353|   353|	})
   354|   354|   354|   354|   354|	return ListGroupsResponse{Memberships: memberships}, nil
   355|   355|   355|   355|   355|}
   356|   356|   356|   356|   356|
   357|   357|   357|   357|   357|func (s *MemoryStore) CreateInvite(ctx context.Context, player Player, groupID string) (Invite, bool, error) {
   358|   358|   358|   358|   358|	s.mu.Lock()
   359|   359|   359|   359|   359|	defer s.mu.Unlock()
   360|   360|   360|   360|   360|
   361|   361|   361|   361|   361|	if _, ok := s.memberships[groupID][player.ID]; !ok {
   362|   362|   362|   362|   362|		return Invite{}, false, nil
   363|   363|   363|   363|   363|	}
   364|   364|   364|   364|   364|
   365|   365|   365|   365|   365|	s.inviteNumber++
   366|   366|   366|   366|   366|	token, err := randomToken("invite_token")
   367|   367|   367|   367|   367|	if err != nil {
   368|   368|   368|   368|   368|		return Invite{}, false, err
   369|   369|   369|   369|   369|	}
   370|   370|   370|   370|   370|	invite := Invite{
   371|   371|   371|   371|   371|		ID:        stableID("invite", groupID+":"+strconv.Itoa(s.inviteNumber)),
   372|   372|   372|   372|   372|		GroupID:   groupID,
   373|   373|   373|   373|   373|		Token:     token,
   374|   374|   374|   374|   374|		CreatedBy: player.ID,
   375|   375|   375|   375|   375|		ExpiresAt: s.now().Add(7 * 24 * time.Hour).UTC(),
   376|   376|   376|   376|   376|	}
   377|   377|   377|   377|   377|	s.invites[invite.Token] = memoryInvite{Invite: invite}
   378|   378|   378|   378|   378|	return invite, true, nil
   379|   379|   379|   379|   379|}
   380|   380|   380|   380|   380|
   381|   381|   381|   381|   381|func (s *MemoryStore) AcceptInvite(ctx context.Context, player Player, token string) (GroupHomeResponse, InviteAcceptStatus, error) {
   382|   382|   382|   382|   382|	s.mu.Lock()
   383|   383|   383|   383|   383|	defer s.mu.Unlock()
   384|   384|   384|   384|   384|
   385|   385|   385|   385|   385|	invite, ok := s.invites[token]
   386|   386|   386|   386|   386|	if !ok {
   387|   387|   387|   387|   387|		return GroupHomeResponse{}, InviteInvalid, nil
   388|   388|   388|   388|   388|	}
   389|   389|   389|   389|   389|	if invite.UsedBy != "" {
   390|   390|   390|   390|   390|		return GroupHomeResponse{}, InviteUsed, nil
   391|   391|   391|   391|   391|	}
   392|   392|   392|   392|   392|	if s.now().After(invite.ExpiresAt) {
   393|   393|   393|   393|   393|		return GroupHomeResponse{}, InviteExpired, nil
   394|   394|   394|   394|   394|	}
   395|   395|   395|   395|   395|	group, ok := s.groups[invite.GroupID]
   396|   396|   396|   396|   396|	if !ok {
   397|   397|   397|   397|   397|		return GroupHomeResponse{}, InviteInvalid, nil
   398|   398|   398|   398|   398|	}
   399|   399|   399|   399|   399|	membership := GroupMembership{GroupID: invite.GroupID, PlayerID: player.ID, Role: "Player"}
   400|   400|   400|   400|   400|	if _, ok := s.memberships[invite.GroupID][player.ID]; ok {
   401|   401|   401|   401|   401|		return GroupHomeResponse{}, InviteMember, nil
   402|   402|   402|   402|   402|	}
   403|   403|   403|   403|   403|	s.memberships[invite.GroupID][player.ID] = membership
   404|   404|   404|   404|   404|	invite.UsedBy = player.ID
   405|   405|   405|   405|   405|	s.invites[token] = invite
   406|   406|   406|   406|   406|	return groupHome(group, membership, s.currentSeasonForGroup(invite.GroupID), s.recentPerformedStuntsForGroup(invite.GroupID), s.standingsForGroup(invite.GroupID)), InviteAccepted, nil
   407|   407|   407|   407|   407|}
   408|   408|   408|   408|   408|
   409|   409|   409|   409|   409|func (s *MemoryStore) StartSeason(ctx context.Context, player Player, groupID string, submissionDeadline time.Time, judgingDeadline time.Time) (GroupHomeResponse, bool, error) {
   410|   410|   410|   410|   410|	s.mu.Lock()
   411|   411|   411|   411|   411|	defer s.mu.Unlock()
   412|   412|   412|   412|   412|
   413|   413|   413|   413|   413|	group, ok := s.groups[groupID]
   414|   414|   414|   414|   414|	if !ok {
   415|   415|   415|   415|   415|		return GroupHomeResponse{}, false, nil
   416|   416|   416|   416|   416|	}
   417|   417|   417|   417|   417|	membership, ok := s.memberships[groupID][player.ID]
   418|   418|   418|   418|   418|	if !ok {
   419|   419|   419|   419|   419|		return GroupHomeResponse{}, false, nil
   420|   420|   420|   420|   420|	}
   421|   421|   421|   421|   421|	if s.openSeasonForGroup(groupID) != nil {
   422|   422|   422|   422|   422|		return GroupHomeResponse{}, true, ErrSeasonAlreadyOpen
   423|   423|   423|   423|   423|	}
   424|   424|   424|   424|   424|
   425|   425|   425|   425|   425|	s.seasonNumber++
   426|   426|   426|   426|   426|	season := Season{
   427|   427|   427|   427|   427|		ID:                   stableID("season", groupID+":"+strconv.Itoa(s.seasonNumber)),
   428|   428|   428|   428|   428|		GroupID:              groupID,
   429|   429|   429|   429|   429|		CommissionerPlayerID: player.ID,
   430|   430|   430|   430|   430|		Status:               "Active",
   431|   431|   431|   431|   431|		SubmissionDeadline:   submissionDeadline.UTC(),
   432|   432|   432|   432|   432|		JudgingDeadline:      judgingDeadline.UTC(),
   433|   433|   433|   433|   433|	}
   434|   434|   434|   434|   434|	s.seasons[season.ID] = season
   435|   435|   435|   435|   435|	s.seasonOrder[season.ID] = s.seasonNumber
   436|   436|   436|   436|   436|	return groupHome(group, membership, &season, s.recentPerformedStuntsForGroup(groupID), s.standingsForGroup(groupID)), true, nil
   437|   437|   437|   437|   437|}
   438|   438|   438|   438|   438|
   439|   439|   439|   439|   439|func (s *MemoryStore) CloseSeasonSubmissions(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error) {
   440|   440|   440|   440|   440|	s.mu.Lock()
   441|   441|   441|   441|   441|	defer s.mu.Unlock()
   442|   442|   442|   442|   442|
   443|   443|   443|   443|   443|	season, ok := s.seasons[seasonID]
   444|   444|   444|   444|   444|	if !ok {
   445|   445|   445|   445|   445|		return GroupHomeResponse{}, false, ErrSeasonNotFound
   446|   446|   446|   446|   446|	}
   447|   447|   447|   447|   447|	group, ok := s.groups[season.GroupID]
   448|   448|   448|   448|   448|	if !ok {
   449|   449|   449|   449|   449|		return GroupHomeResponse{}, false, ErrSeasonNotFound
   450|   450|   450|   450|   450|	}
   451|   451|   451|   451|   451|	membership, ok := s.memberships[season.GroupID][player.ID]
   452|   452|   452|   452|   452|	if !ok {
   453|   453|   453|   453|   453|		return GroupHomeResponse{}, false, nil
   454|   454|   454|   454|   454|	}
   455|   455|   455|   455|   455|	if player.ID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
   456|   456|   456|   456|   456|		return GroupHomeResponse{}, false, nil
   457|   457|   457|   457|   457|	}
   458|   458|   458|   458|   458|	if season.Status == "Active" {
   459|   459|   459|   459|   459|		fromStatus := season.Status
   460|   460|   460|   460|   460|		season.Status = "Judging Grace Period"
   461|   461|   461|   461|   461|		s.seasons[season.ID] = season
   462|   462|   462|   462|   462|		s.recordSeasonHistory(season.ID, "Submissions Closed", player.ID, membership.Role, player.ID != season.CommissionerPlayerID, fromStatus, season.Status)
   463|   463|   463|   463|   463|	}
   464|   464|   464|   464|   464|	return groupHome(group, membership, s.currentSeasonForGroup(season.GroupID), s.recentPerformedStuntsForGroup(season.GroupID), s.standingsForGroup(season.GroupID)), true, nil
   465|   465|   465|   465|   465|}
   466|   466|   466|   466|   466|
   467|   467|   467|   467|   467|func (s *MemoryStore) FinalizeSeason(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error) {
   468|   468|   468|   468|   468|	s.mu.Lock()
   469|   469|   469|   469|   469|	defer s.mu.Unlock()
   470|   470|   470|   470|   470|
   471|   471|   471|   471|   471|	season, ok := s.seasons[seasonID]
   472|   472|   472|   472|   472|	if !ok {
   473|   473|   473|   473|   473|		return GroupHomeResponse{}, false, ErrSeasonNotFound
   474|   474|   474|   474|   474|	}
   475|   475|   475|   475|   475|	group, ok := s.groups[season.GroupID]
   476|   476|   476|   476|   476|	if !ok {
   477|   477|   477|   477|   477|		return GroupHomeResponse{}, false, ErrSeasonNotFound
   478|   478|   478|   478|   478|	}
   479|   479|   479|   479|   479|	membership, ok := s.memberships[season.GroupID][player.ID]
   480|   480|   480|   480|   480|	if !ok {
   481|   481|   481|   481|   481|		return GroupHomeResponse{}, false, nil
   482|   482|   482|   482|   482|	}
   483|   483|   483|   483|   483|	if player.ID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
   484|   484|   484|   484|   484|		return GroupHomeResponse{}, false, nil
   485|   485|   485|   485|   485|	}
   486|   486|   486|   486|   486|	if season.Status != "Finalized" {
   487|   487|   487|   487|   487|		fromStatus := season.Status
   488|   488|   488|   488|   488|		season.Status = "Finalized"
   489|   489|   489|   489|   489|		s.seasons[season.ID] = season
   490|   490|   490|   490|   490|		s.finalizeSeasonStunts(season.ID)
   491|   491|   491|   491|   491|		s.recordSeasonHistory(season.ID, "Season Finalized", player.ID, membership.Role, player.ID != season.CommissionerPlayerID, fromStatus, season.Status)
   492|   492|   492|   492|   492|	}
   493|   493|   493|   493|   493|	return groupHome(group, membership, s.currentSeasonForGroup(season.GroupID), s.recentPerformedStuntsForGroup(season.GroupID), s.standingsForGroup(season.GroupID)), true, nil
   494|   494|   494|   494|   494|}
   495|   495|   495|   495|   495|
   496|   496|   496|   496|   496|func (s *MemoryStore) SeasonHistory(ctx context.Context, player Player, seasonID string) (SeasonHistoryResponse, bool, error) {
   497|   497|   497|   497|   497|	s.mu.Lock()
   498|   498|   498|   498|   498|	defer s.mu.Unlock()
   499|   499|   499|   499|   499|
   500|   500|   500|   500|   500|	season, ok := s.seasons[seasonID]
   501|