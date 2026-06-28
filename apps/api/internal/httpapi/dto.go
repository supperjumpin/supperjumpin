package httpapi

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
