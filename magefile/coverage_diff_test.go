package main

import (
	"strings"
	"testing"
)

func TestCoverageComment(t *testing.T) {
	current := map[string]*CoverageReport{
		"api": {Total: 82.5},
		"bot": {Total: 71.0},
	}
	baseline := map[string]*CoverageReport{
		"api": {Total: 80.0},
		"bot": {Total: 72.0},
	}

	got := CoverageComment(current, baseline)

	wants := []string{
		"### Coverage Report",
		"| **Go API** | 80.0% | 82.5% | +2.5% ✅ |",
		"| **Discord Bot** | 72.0% | 71.0% | -1.0% 🔻 |",
		"> _Non-blocking coverage report._",
	}

	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Fatalf("CoverageComment() missing %q in:\n%s", want, got)
		}
	}
}

func TestCoverageCommentSkipsUnknownScopes(t *testing.T) {
	got := CoverageComment(map[string]*CoverageReport{"api": {Total: 55.0}}, map[string]*CoverageReport{})
	if strings.Contains(got, "Mobile") || strings.Contains(got, "api-client") {
		t.Fatalf("CoverageComment() should not mention deleted scopes:\n%s", got)
	}
	if !strings.Contains(got, "| **Go API** | — | 55.0% | — |") {
		t.Fatalf("CoverageComment() missing Go API row:\n%s", got)
	}
}
