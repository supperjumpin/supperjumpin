//go:build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
)

type suite struct {
	Key          string
	Label        string
	ModuleDir    string
	CoverProfile string
	ReportPath   string
}

// Test runs API and bot tests; pass -coverage to write coverage artifacts.
func Test(coverage *bool) error {
	enabled := coverage != nil && *coverage
	testDatabaseURL := testDatabaseURLFromEnv(os.Getenv)
	suites := []suite{
		{
			Key:          "api",
			Label:        "Go API",
			ModuleDir:    repoPath("apps", "api"),
			CoverProfile: repoPath("coverage", "api.coverprofile"),
			ReportPath:   repoPath("coverage", "api-report.json"),
		},
		{
			Key:          "bot",
			Label:        "Discord Bot",
			ModuleDir:    repoPath("apps", "bot-discord"),
			CoverProfile: repoPath("coverage", "bot.coverprofile"),
			ReportPath:   repoPath("coverage", "bot-report.json"),
		},
	}

	if enabled {
		if err := os.MkdirAll(repoPath("coverage"), 0o755); err != nil {
			return fmt.Errorf("create coverage dir: %w", err)
		}
	}

	for _, suite := range suites {
		if suite.Key == "api" {
			if err := prepareAPITestDatabase(testDatabaseURL); err != nil {
				return err
			}
		}
		coverProfile := ""
		if enabled {
			coverProfile = suite.CoverProfile
		}
		env := []string(nil)
		if suite.Key == "api" {
			env = []string{"SUPPERJUMPIN_TEST_DATABASE_URL=" + testDatabaseURL}
		}
		if err := runner.Run(goTestCommandWithEnv(suite.ModuleDir, coverProfile, nil, env)); err != nil {
			return err
		}
		if !enabled {
			continue
		}
		if err := writeCoverageArtifacts(suite); err != nil {
			return err
		}
	}

	return nil
}

func prepareAPITestDatabase(testDatabaseURL string) error {
	dbName, err := ParseDatabaseName(testDatabaseURL)
	if err != nil {
		return err
	}
	allowUnsafe := os.Getenv("SUPPERJUMPIN_TEST_ALLOW_UNSAFE_RESET") == "1"
	if !IsSafeToReset(dbName, allowUnsafe) {
		return fmt.Errorf("refusing to reset database %q because it does not end with _test", dbName)
	}
	localDocker := os.Getenv("SUPPERJUMPIN_TEST_DATABASE_URL") == ""
	if localDocker {
		db := DB{}
		if err := db.Up(); err != nil {
			return err
		}
		if err := waitForPostgresReady(defaultWaitTimeout, defaultWaitInterval); err != nil {
			return err
		}
	}
	commands, err := dbResetCommands(testDatabaseURL, localDocker, migrateBinDirFromEnv(os.Getenv))
	if err != nil {
		return err
	}
	return runAll(commands...)
}

func writeCoverageArtifacts(s suite) error {
	output, err := runner.Output(goCoverFuncCommand(s.ModuleDir, s.CoverProfile))
	if output != "" {
		fmt.Print(output)
	}
	if err != nil {
		return err
	}
	total, err := parseTotalCoverage(output)
	if err != nil {
		return err
	}
	if err := writeCoverageReportFile(s.ReportPath, CoverageReport{Total: total}); err != nil {
		return err
	}
	return appendCoverageSummary(s.Label, total, os.Getenv("GITHUB_STEP_SUMMARY"))
}

var _ mg.Namespace
