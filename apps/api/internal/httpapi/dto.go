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
