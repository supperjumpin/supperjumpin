//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
)

type DB mg.Namespace

func (DB) Up() error {
	spec, err := dbLifecycleCommand("up")
	if err != nil {
		return err
	}
	return runner.Run(spec)
}

func (DB) Down() error {
	spec, err := dbLifecycleCommand("down")
	if err != nil {
		return err
	}
	return runner.Run(spec)
}

func (DB) Migrate() error {
	return runner.Run(dbMigrateCommand(DefaultDevelopmentDatabaseURL, migrateBinDirFromEnv(os.Getenv)))
}

func (DB) Reset() error {
	db := DB{}
	if err := db.Up(); err != nil {
		return err
	}
	if err := waitForPostgresReady(defaultWaitTimeout, defaultWaitInterval); err != nil {
		return err
	}
	commands, err := dbResetCommands(DefaultDevelopmentDatabaseURL, true, migrateBinDirFromEnv(os.Getenv))
	if err != nil {
		return err
	}
	return runAll(commands...)
}
