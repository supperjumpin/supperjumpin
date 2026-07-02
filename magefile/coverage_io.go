package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadCoverageReport(path string) (*CoverageReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read coverage report: %w", err)
	}
	var report CoverageReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse coverage report: %w", err)
	}
	return &report, nil
}

func loadCoverageReports(dir string) (map[string]*CoverageReport, error) {
	reports := make(map[string]*CoverageReport)
	for _, scope := range []string{"api", "bot"} {
		report, err := loadCoverageReport(filepath.Join(dir, scope+"-report.json"))
		if err != nil {
			return nil, err
		}
		reports[scope] = report
	}
	return reports, nil
}

func writeCoverageReportFile(path string, report CoverageReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create coverage artifact dir: %w", err)
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal coverage report: %w", err)
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

func appendCoverageSummary(label string, total float64, summaryPath string) error {
	if summaryPath == "" {
		return nil
	}
	content := fmt.Sprintf("### %s coverage\n\n- total: %.1f%%\n", label, total)
	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open github step summary: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
