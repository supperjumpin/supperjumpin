package main

func apiDevCommand(getenv func(string) string) CommandSpec {
	databaseURL := valueOrDefault(getenv("SUPPERJUMPIN_DATABASE_URL"), DefaultDevelopmentDatabaseURL)
	devToken := valueOrDefault(getenv("SUPPERJUMPIN_DEV_AUTH_TOKEN"), "dev-token")
	adapterToken := valueOrDefault(getenv("SUPPERJUMPIN_ADAPTER_TOKEN"), devToken)
	return CommandSpec{
		Name: "go",
		Dir:  repoPath("apps", "api"),
		Args: []string{"run", "./cmd/api"},
		Env: []string{
			"SUPPERJUMPIN_DATABASE_URL=" + databaseURL,
			"SUPPERJUMPIN_DEV_AUTH_TOKEN=" + devToken,
			"SUPPERJUMPIN_ADAPTER_TOKEN=" + adapterToken,
		},
	}
}

func botDevCommand(getenv func(string) string) CommandSpec {
	botToken := valueOrDefault(getenv("SUPPERJUMPIN_BOT_TOKEN"), "Bot dev-placeholder-token")
	adapterToken := valueOrDefault(getenv("SUPPERJUMPIN_ADAPTER_TOKEN"), "dev-token")
	apiBaseURL := valueOrDefault(getenv("SUPPERJUMPIN_API_BASE_URL"), "http://localhost:8080")
	return CommandSpec{
		Name: "go",
		Dir:  repoPath("apps", "bot-discord"),
		Args: []string{"run", "./cmd/bot"},
		Env: []string{
			"SUPPERJUMPIN_BOT_TOKEN=" + botToken,
			"SUPPERJUMPIN_ADAPTER_TOKEN=" + adapterToken,
			"SUPPERJUMPIN_API_BASE_URL=" + apiBaseURL,
		},
	}
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
