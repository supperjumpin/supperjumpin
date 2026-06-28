package httpapi

type Player struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type Community struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type MeResponse struct {
	Player    Player    `json:"player"`
	Community Community `json:"community"`
}

type PromptPackDTO struct {
	ID          string      `json:"id"`
	DisplayName string      `json:"displayName"`
	Description string      `json:"description"`
	Prompts     []PromptDTO `json:"prompts"`
}

type PromptDTO struct {
	ID       string `json:"id"`
	Copy     string `json:"copy"`
	Theme    string `json:"theme"`
	CostTier string `json:"costTier"`
}

type PromptCatalogResponse struct {
	Packs []PromptPackDTO `json:"packs"`
}

type RevealTimeframeDTO struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	DurationHours int    `json:"durationHours"`
}

type RevealTimeframesResponse struct {
	Timeframes []RevealTimeframeDTO `json:"timeframes"`
}

type RoundDTO struct {
	ID          string      `json:"id"`
	CommunityID string      `json:"communityId"`
	PromptID    string      `json:"promptId"`
	Status      string      `json:"status"`
	RevealBy    string      `json:"revealBy"`
	CreatedBy   string      `json:"createdBy"`
	CreatedAt   string      `json:"createdAt"`
	Prompt      *PromptDTO  `json:"prompt,omitempty"`
}

type StartRoundRequest struct {
	CommunityID       string `json:"communityId"`
	PromptID          string `json:"promptId,omitempty"`
	RevealTimeframeID string `json:"revealTimeframeId"`
}

type StartRoundResponse struct {
	Round RoundDTO `json:"round"`
}

type CommitRequest struct {
	// intentionally empty — the player is from auth, round from path
}

type CommitResponse struct {
	CommitID string `json:"commitId"`
}

type SubmitJumpRequest struct {
	Caption      string   `json:"caption"`
	EvidenceURLs []string `json:"evidenceUrls"`
}

type SubmitJumpResponse struct {
	Jump JumpDTO `json:"jump"`
}

type JumpDTO struct {
	ID                 string   `json:"id"`
	RoundID            string   `json:"roundId"`
	PlayerID           string   `json:"playerId"`
	Caption            string   `json:"caption,omitempty"`
	EvidenceURLs       []string `json:"evidenceUrls,omitempty"`
	SubmittedAt        string   `json:"submittedAt"`
	SealedViewer       bool     `json:"sealedViewer"`
	PlayerHasCommitted  bool    `json:"playerHasCommitted"`
	PlayerHasSubmitted  bool    `json:"playerHasSubmitted"`
}

type ListJumpsResponse struct {
	Jumps          []JumpDTO `json:"jumps"`
	CommitCount    int       `json:"commitCount"`
	SubmissionCount int      `json:"submissionCount"`
}

type GetJumpResponse struct {
	Jump JumpDTO `json:"jump"`
}

type RevealRoundResponse struct {
	Round    RoundDTO `json:"round"`
	Revealed bool     `json:"revealed"`
}

type StampDTO struct {
	ID     string `json:"id"`
	Stance string `json:"stance"`
	Label  string `json:"label"`
	Glyph  string `json:"glyph,omitempty"`
	Copy   string `json:"copy,omitempty"`
}

type StampCatalogResponse struct {
	Stamps []StampDTO `json:"stamps"`
}

type ApplyReactionRequest struct {
	StampID string `json:"stampId"`
}

type ReactionDTO struct {
	ID        string `json:"id"`
	StampID   string `json:"stampId"`
	JumpID    string `json:"jumpId"`
	PlayerID  string `json:"playerId"`
	CreatedAt string `json:"createdAt"`
}

type ApplyReactionResponse struct {
	Reaction ReactionDTO `json:"reaction"`
}

type PostCommentRequest struct {
	Body string `json:"body"`
}

type CommentDTO struct {
	ID        string `json:"id"`
	RoundID   string `json:"roundId"`
	JumpID    string `json:"jumpId,omitempty"`
	PlayerID  string `json:"playerId"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

type PostCommentResponse struct {
	Comment CommentDTO `json:"comment"`
}

type ListCommentsResponse struct {
	Comments []CommentDTO `json:"comments"`
}

type LoreEntryDTO struct {
	JumpID       string         `json:"jumpId"`
	RoundID      string         `json:"roundId"`
	JumpCaption  string         `json:"jumpCaption"`
	JumpPlayerID string         `json:"jumpPlayerId"`
	StampCounts  map[string]int `json:"stampCounts"`
	TotalStamps  int            `json:"totalStamps"`
}

type CommunityLoreResponse struct {
	Entries []LoreEntryDTO `json:"entries"`
}

type RecapJumpEntryDTO struct {
	JumpID       string         `json:"jumpId"`
	PlayerID     string         `json:"playerId"`
	Caption      string         `json:"caption"`
	EvidenceURLs []string       `json:"evidenceUrls"`
	SubmittedAt  string         `json:"submittedAt"`
	StampCounts  map[string]int `json:"stampCounts"`
	TotalStamps  int            `json:"totalStamps"`
}

type GhostJumperDTO struct {
	PlayerID    string `json:"playerId"`
	CommittedAt string `json:"committedAt"`
}

type NextRoundHookDTO struct {
	ActiveRoundID string `json:"activeRoundId,omitempty"`
	PromptID      string `json:"promptId,omitempty"`
}

type StandoutStampDTO struct {
	JumpID string `json:"jumpId"`
	Stance string `json:"stance"`
	Count  int    `json:"count"`
}

type RecapResponse struct {
	RoundID          string               `json:"roundId"`
	CommunityID      string               `json:"communityId"`
	PromptID         string               `json:"promptId"`
	Status           string               `json:"status"`
	RevealBy         string               `json:"revealBy"`
	CreatedBy        string               `json:"createdBy"`
	CreatedAt        string               `json:"createdAt"`
	Jumps            []RecapJumpEntryDTO  `json:"jumps"`
	Comments         []CommentDTO         `json:"comments"`
	GhostJumpers     []GhostJumperDTO     `json:"ghostJumpers"`
	Lore             []LoreEntryDTO       `json:"lore"`
	NextRoundHook    NextRoundHookDTO     `json:"nextRoundHook"`
	StandoutStamps   []StandoutStampDTO   `json:"standoutStamps"`
	StandoutComments []CommentDTO         `json:"standoutComments"`
}
