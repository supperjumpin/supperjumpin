package main

import "testing"

func TestAgentAttemptFromEnv(t *testing.T) {
	attempt, err := agentAttemptFromEnv(func(key string) string {
		switch key {
		case "AGENT_TASK_ID":
			return "issue-123"
		case "AGENT_ATTEMPT_ID":
			return "attempt-1"
		case "AGENT_TASK_CLASS":
			return "qa"
		case "GITHUB_SHA":
			return "abc123"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if attempt.Artifacts != "artifacts/agents/issue-123/attempt-1" {
		t.Fatalf("Artifacts = %q", attempt.Artifacts)
	}
	if attempt.TaskClass != "qa" || attempt.SourceRevision != "abc123" {
		t.Fatalf("attempt = %#v", attempt)
	}
}

func TestAgentAttemptFromEnvRejectsUnsafePath(t *testing.T) {
	_, err := agentAttemptFromEnv(func(key string) string {
		if key == "AGENT_TASK_ID" {
			return "../outside"
		}
		return "attempt-1"
	})
	if err == nil {
		t.Fatal("agentAttemptFromEnv() error = nil, want unsafe task ID error")
	}
}
