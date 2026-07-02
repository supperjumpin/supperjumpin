//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

type CI mg.Namespace

func (CI) All() {
	mg.SerialDeps(CI{}.Lint, mg.F(Test, boolPtr(true)), CI{}.Build)
}

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

func (CI) Test() error {
	value := true
	return Test(&value)
}

func (CI) Build() error {
	build := Build{}
	if err := build.API(); err != nil {
		return err
	}
	return build.Bot()
}

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
