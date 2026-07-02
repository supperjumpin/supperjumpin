//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
)

type Dev mg.Namespace

// API runs the API service against an existing database.
func (Dev) API() error {
	return runner.Run(apiDevCommand(os.Getenv))
}

// Bot runs the Discord bot against the local API.
func (Dev) Bot() error {
	return runner.Run(botDevCommand(os.Getenv))
}
