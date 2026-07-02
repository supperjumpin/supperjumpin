//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

type CI mg.Namespace

// All runs lint, coverage tests, and Docker image builds.
func (CI) All() {
	mg.SerialDeps(CI{}.Lint, mg.F(Test, boolPtr(true)), CI{}.Build)
}

// Lint runs Go vet and verifies generated SQLC output is current.
func (CI) Lint() error {
	generate := Generate{}
	if err := runner.Run(goVetCommand(repoPath("apps", "api"))); err != nil {
		return err
	}
	if err := runner.Run(goVetCommand(repoPath("apps", "bot-discord"))); err != nil {
		return err
	}
	if err := generate.SQLC(); err != nil {
		return err
	}
	return runner.Run(gitDiffExitCodeCommand("apps/api/internal/db"))
}

// Test runs the CI test path with coverage enabled.
func (CI) Test() error {
	value := true
	return Test(&value)
}

// Build builds both service Docker images.
func (CI) Build() error {
	build := Build{}
	if err := build.API(); err != nil {
		return err
	}
	return build.Bot()
}

// Comment prints a markdown coverage comparison for PR automation.
func (CI) Comment(currentDir string, baselineDir string) error {
	current, err := loadCoverageReports(currentDir)
	if err != nil {
		return err
	}
	baseline, err := loadCoverageReports(baselineDir)
	if err != nil {
		return err
	}
	fmt.Print(CoverageComment(current, baseline))
	return nil
}

func boolPtr(v bool) *bool { return &v }
