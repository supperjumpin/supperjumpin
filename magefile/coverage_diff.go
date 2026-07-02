package main

import (
	"fmt"
	"strings"
)

type CoverageReport struct {
	Total float64 `json:"total"`
}

func CoverageComment(current, baseline map[string]*CoverageReport) string {
	type scope struct {
		key   string
		label string
	}
	scopes := []scope{{key: "api", label: "Go API"}, {key: "bot", label: "Discord Bot"}}

	lines := []string{
		"### Coverage Report",
		"",
		"| Scope | Baseline | Current | Change |",
		"|---|---|---|---|",
	}

	for _, scope := range scopes {
		cur := current[scope.key]
		base := baseline[scope.key]
		if cur == nil && base == nil {
			continue
		}
		if cur != nil && base != nil {
			delta := cur.Total - base.Total
			lines = append(lines, fmt.Sprintf("| **%s** | %s | %s | %s%s |", scope.label, pct(base.Total), pct(cur.Total), signedPct(delta), deltaIcon(delta)))
			continue
		}
		lines = append(lines, fmt.Sprintf("| **%s** | %s | %s | — |", scope.label, nullablePct(base), nullablePct(cur)))
	}

	lines = append(lines, "", "> _Non-blocking coverage report._")
	return strings.Join(lines, "\n") + "\n"
}

func pct(v float64) string {
	return fmt.Sprintf("%.1f%%", v)
}

func nullablePct(report *CoverageReport) string {
	if report == nil {
		return "—"
	}
	return pct(report.Total)
}

func signedPct(delta float64) string {
	if delta > 0 {
		return fmt.Sprintf("+%.1f%%", delta)
	}
	return fmt.Sprintf("%.1f%%", delta)
}

func deltaIcon(delta float64) string {
	if delta < -0.5 {
		return " 🔻"
	}
	if delta > 0.5 {
		return " ✅"
	}
	return ""
}
