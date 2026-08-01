package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type agentAttempt struct {
	TaskID         string `json:"task_id"`
	AttemptID      string `json:"attempt_id"`
	TaskClass      string `json:"task_class"`
	SourceRevision string `json:"source_revision"`
	StartedAt      string `json:"started_at"`
	CompletedAt    string `json:"completed_at"`
	Result         string `json:"result"`
	FailureClass   string `json:"failure_class,omitempty"`
	FailureMessage string `json:"failure_message,omitempty"`
	Artifacts      string `json:"artifacts"`
}

func agentAttemptFromEnv(getenv func(string) string) (agentAttempt, error) {
	taskID := getenv("AGENT_TASK_ID")
	attemptID := getenv("AGENT_ATTEMPT_ID")
	sourceRevision := valueOrDefault(getenv("AGENT_SOURCE_REVISION"), getenv("GITHUB_SHA"))
	if !safePathComponent(taskID) {
		return agentAttempt{}, fmt.Errorf("AGENT_TASK_ID must contain only letters, digits, dot, underscore, or hyphen")
	}
	if !safePathComponent(attemptID) {
		return agentAttempt{}, fmt.Errorf("AGENT_ATTEMPT_ID must contain only letters, digits, dot, underscore, or hyphen")
	}
	if strings.TrimSpace(sourceRevision) == "" {
		return agentAttempt{}, fmt.Errorf("AGENT_SOURCE_REVISION or GITHUB_SHA is required")
	}

	return agentAttempt{
		TaskID:         taskID,
		AttemptID:      attemptID,
		TaskClass:      valueOrDefault(getenv("AGENT_TASK_CLASS"), "implementation"),
		SourceRevision: strings.TrimSpace(sourceRevision),
		Artifacts:      filepath.ToSlash(filepath.Join("artifacts", "agents", taskID, attemptID)),
	}, nil
}

func safePathComponent(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '.' && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func writeAgentAttempt(attempt agentAttempt) error {
	dir := repoPath(attempt.Artifacts)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create agent artifacts: %w", err)
	}
	data, err := json.MarshalIndent(attempt, "", "  ")
	if err != nil {
		return fmt.Errorf("encode agent result: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(dir, "result.json"), data, 0o644); err != nil {
		return fmt.Errorf("write agent result: %w", err)
	}
	return nil
}

func finishAgentAttempt(attempt agentAttempt, err error) error {
	attempt.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	if err == nil {
		attempt.Result = "passed"
	} else {
		attempt.Result = "failed"
		if attempt.FailureClass == "" {
			attempt.FailureClass = "verification"
		}
		attempt.FailureMessage = err.Error()
	}
	if writeErr := writeAgentAttempt(attempt); writeErr != nil {
		return writeErr
	}
	return err
}
