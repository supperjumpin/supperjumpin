//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
)

type Generate mg.Namespace

// SQLC regenerates the API query layer from SQL files.
func (Generate) SQLC() error {
	return runner.Run(sqlcGenerateCommand(sqlcBinDirFromEnv(os.Getenv)))
}

func sqlcBinDirFromEnv(getenv func(string) string) string {
	if dir := getenv("SUPPERJUMPIN_SQLC_BIN_DIR"); dir != "" {
		return dir
	}
	return "bin"
}
