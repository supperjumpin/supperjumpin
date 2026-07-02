//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
)

type DB mg.Namespace

// Up starts the local Docker Compose Postgres service.
func (DB) Up() error {
	spec, err := dbLifecycleCommand("up")
	if err != nil {
		return err
	}
	return runner.Run(spec)
}

// Down stops the local Docker Compose Postgres service.
func (DB) Down() error {
	spec, err := dbLifecycleCommand("down")
	if err != nil {
		return err
	}
	return runner.Run(spec)
}

// Migrate applies API migrations to the local development database.
func (DB) Migrate() error {
	return runner.Run(dbMigrateCommand(DefaultDevelopmentDatabaseURL, migrateBinDirFromEnv(os.Getenv)))
}

// Reset recreates the local development database and reapplies migrations.
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
