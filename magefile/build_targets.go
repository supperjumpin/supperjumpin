//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
)

type Build mg.Namespace
type Generate mg.Namespace
type Dev mg.Namespace

func (Build) API() error {
	return runner.Run(buildImageCommand("Dockerfile.api", "supperjumpin-api:dev"))
}

func (Build) Bot() error {
	return runner.Run(buildImageCommand("Dockerfile.bot", "supperjumpin-bot:dev"))
}

func (Generate) SQLC() error {
	return runner.Run(sqlcGenerateCommand(sqlcBinDirFromEnv(os.Getenv)))
}

func (Dev) API() error {
	return runner.Run(apiDevCommand(os.Getenv))
}

func (Dev) Bot() error {
	return runner.Run(botDevCommand(os.Getenv))
}

func sqlcBinDirFromEnv(getenv func(string) string) string {
	if dir := getenv("SUPPERJUMPIN_SQLC_BIN_DIR"); dir != "" {
		return dir
	}
	return "bin"
}
