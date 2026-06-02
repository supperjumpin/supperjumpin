package httpapi

import "time"

type Account struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Player struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type MeResponse struct {
	Account Account `json:"account"`
	Player  Player  `json:"player"`
}

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GroupMembership struct {
	GroupID  string `json:"groupId"`
	PlayerID string `json:"playerId"`
	Role     string `json:"role"`
}

type Season struct {
	ID                   string    `json:"id"`
	GroupID              string    `json:"groupId"`
	CommissionerPlayerID string    `json:"commissionerPlayerId"`
	Status               string    `json:"status"`
	SubmissionDeadline   time.Time `json:"submissionDeadline"`
	JudgingDeadline      time.Time `json:"judgingDeadline"`
}

type SeasonHistoryEntry struct {
	ID            string    `json:"id"`
	SeasonID      string    `json:"seasonId"`
	Action        string    `json:"action"`
	ActorPlayerID string    `json:"actorPlayerId"`
	ActorRole     string    `json:"actorRole"`
	Override      bool      `json:"override"`
	FromStatus    string    `json:"fromStatus"`
	ToStatus      string    `json:"toStatus"`
	CreatedAt     time.Time `json:"createdAt"`
}

type SeasonHistoryResponse struct {
	Entries []SeasonHistoryEntry `json:"entries"`
}

type GroupHomeResponse struct {
	Group        Group               `json:"group"`
	Membership   GroupMembership     `json:"membership"`
	ActiveSeason *Season             `json:"activeSeason"`
	RecentJumps  []PerformedJumpView `json:"recentJumps"`
	Standings    []StandingEntry     `json:"standings"`
}

type StandingEntry struct {
	Player      Player `json:"player"`
	SeasonScore int    `json:"seasonScore"`
	JudgedJumps int    `json:"judgedJumps"`
}

type PerformedJumpView struct {
	Jump      Jump      `json:"jump"`
	Performer Player    `json:"performer"`
	Evidence  Evidence  `json:"evidence"`
	Disputes  []Dispute `json:"disputes"`
}

type Jump struct {
	ID                   string     `json:"id"`
	GroupID              string     `json:"groupId"`
	PlayerID             string     `json:"playerId"`
	SeasonID             *string    `json:"seasonId"`
	Status               string     `json:"status"`
	Source               string     `json:"source"`
	Destination          string     `json:"destination"`
	Food                 string     `json:"food"`
	OffSeason            bool       `json:"offSeason"`
	FinalScore           *int       `json:"finalScore"`
	OpenFinalScore       *int       `json:"openFinalScore"`
	SeasonFinalScore     *int       `json:"seasonFinalScore"`
	RemovedAt            *time.Time `json:"-"`
	GracePeriodExpiresAt time.Time  `json:"gracePeriodExpiresAt"`
	CreatedAt            time.Time  `json:"createdAt"`
}

type EvidenceUploadAuthorization struct {
	ID             string            `json:"id"`
	JumpID         string            `json:"jumpId"`
	UploadURL      string            `json:"uploadUrl"`
	UploadMethod   string            `json:"uploadMethod"`
	UploadHeaders  map[string]string `json:"uploadHeaders"`
	MediaObjectKey string            `json:"mediaObjectKey"`
	ExpiresAt      time.Time         `json:"expiresAt"`
}

type Evidence struct {
	ID             string    `json:"id"`
	JumpID         string    `json:"jumpId"`
	Caption        string    `json:"caption"`
	MediaObjectKey string    `json:"mediaObjectKey"`
	CreatedAt      time.Time `json:"createdAt"`
}

type EvidenceSubmission struct {
	Jump     Jump     `json:"jump"`
	Evidence Evidence `json:"evidence"`
}

type Judgment struct {
	ID             string `json:"id"`
	JumpID         string `json:"jumpId"`
	PlayerID       string `json:"playerId"`
	GuestSessionID string `json:"guestSessionId,omitempty"`
	Provenance     string `json:"provenance,omitempty"`
	Commitment     int    `json:"commitment"`
	Transgression  int    `json:"transgression"`
	Creativity     int    `json:"creativity"`
	Presentation   int    `json:"presentation"`
}

type GuestSession struct {
	ID            string `json:"id"`
	JudgmentCount int    `json:"judgmentCount"`
	CreatedAt     int64  `json:"createdAt"`
}

type Dispute struct {
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

type DisputeResolution struct {
	Jump    Jump    `json:"jump"`
	Dispute Dispute `json:"dispute"`
}

type GroupMembershipSummary struct {
	Group      Group           `json:"group"`
	Membership GroupMembership `json:"membership"`
}

type ListGroupsResponse struct {
	Memberships []GroupMembershipSummary `json:"memberships"`
}

type Invite struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"groupId"`
	Token     string    `json:"token"`
	CreatedBy string    `json:"createdBy"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type InviteAcceptStatus string

const (
	InviteAccepted InviteAcceptStatus = "accepted"
	InviteInvalid  InviteAcceptStatus = "invalid"
	InviteUsed     InviteAcceptStatus = "used"
	InviteExpired  InviteAcceptStatus = "expired"
	InviteMember   InviteAcceptStatus = "member"
)

// --- Public Feed / Read Path DTOs ---

type JumpCard struct {
	ID                   string         `json:"id"`
	PerformerName        string         `json:"performerName"`
	PerformerID          string         `json:"performerId"`
	Source               string         `json:"source"`
	Destination          string         `json:"destination"`
	Food                 string         `json:"food"`
	Caption              string         `json:"caption"`
	MediaObjectKey       string         `json:"mediaObjectKey"`
	Status               string         `json:"status"`
	GracePeriodExpiresAt time.Time      `json:"gracePeriodExpiresAt"`
	RunningAverage       float64        `json:"runningAverage"`
	JudgmentCount        int            `json:"judgmentCount"`
	CreatedAt            time.Time      `json:"createdAt"`
	ViewerContext        *ViewerContext `json:"viewerContext,omitempty"`
}

type JumpDetail struct {
	ID                   string         `json:"id"`
	PerformerName        string         `json:"performerName"`
	PerformerID          string         `json:"performerId"`
	Source               string         `json:"source"`
	Destination          string         `json:"destination"`
	Food                 string         `json:"food"`
	Caption              string         `json:"caption"`
	MediaObjectKey       string         `json:"mediaObjectKey"`
	Status               string         `json:"status"`
	GracePeriodExpiresAt time.Time      `json:"gracePeriodExpiresAt"`
	RunningAverage       float64        `json:"runningAverage"`
	JudgmentCount        int            `json:"judgmentCount"`
	CreatedAt            time.Time      `json:"createdAt"`
	RemovedAt            *time.Time     `json:"-"`
	FinalScore           *int           `json:"finalScore,omitempty"`
	Disputes             []Dispute      `json:"disputes,omitempty"`
	ViewerContext        *ViewerContext `json:"viewerContext,omitempty"`
}

type JumpTombstone struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	RemovedAt string `json:"removedAt"`
}

type ViewerContext struct {
	CanJudge          bool       `json:"canJudge"`
	Reason            *string    `json:"reason,omitempty"`
	GracePeriodEndsAt *time.Time `json:"gracePeriodEndsAt,omitempty"`
	HasJudged         bool       `json:"hasJudged"`
}

type PublicFeedResponse struct {
	Jumps      []JumpCard `json:"jumps"`
	NextCursor *string    `json:"nextCursor"`
}
