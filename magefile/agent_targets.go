//go:build mage

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
)

type Agent mg.Namespace

// Verify runs the canonical gates against a runner-provided isolated test database.
func (Agent) Verify() error {
	attempt, err := agentAttemptFromEnv(os.Getenv)
	if err != nil {
		return err
	}
	testDatabaseURL := strings.TrimSpace(os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL"))
	if testDatabaseURL == "" {
		return fmt.Errorf("SUPPERJUMPIN_TEST_DATABASE_URL is required; the runner must provide an isolated database ending in _test")
	}
	dbName, err := ParseDatabaseName(testDatabaseURL)
	if err != nil || !IsSafeToReset(dbName, false) {
		return fmt.Errorf("SUPPERJUMPIN_TEST_DATABASE_URL must name an isolated database ending in _test")
	}
	attempt.StartedAt = time.Now().UTC().Format(time.RFC3339)
	if err := runner.Run(psqlCommand(testDatabaseURL, "SELECT 1;", false)); err != nil {
		attempt.FailureClass = "runner/test_database_unavailable"
		return finishAgentAttempt(attempt, err)
	}
	if err := (CI{}).Lint(); err != nil {
		return finishAgentAttempt(attempt, err)
	}
	return finishAgentAttempt(attempt, (CI{}).Test())
}
