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

type Season struct {
	ID                   string    `json:"id"`
	GroupID              string    `json:"groupId"`
	CommissionerPlayerID string    `json:"commissionerPlayerId"`
	Status               string    `json:"status"`
	SubmissionDeadline   time.Time `json:"submissionDeadline"`
	JudgingDeadline      time.Time `json:"judgingDeadline"`
}

type Jump struct {
	ID                   string     `json:"id"`
	PlayerID             string     `json:"playerId"`
	Status               string     `json:"status"`
	Source               string     `json:"source"`
	Destination          string     `json:"destination"`
	Food                 string     `json:"food"`
	FinalScore           *int       `json:"finalScore"`
	OpenFinalScore       *int       `json:"openFinalScore"`
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
