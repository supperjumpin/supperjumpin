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
