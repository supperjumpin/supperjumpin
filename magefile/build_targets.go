//go:build mage

package main

import (
	"github.com/magefile/mage/mg"
)

type Build mg.Namespace

// API builds the API Docker image as supperjumpin-api:dev.
func (Build) API() error {
	return runner.Run(buildImageCommand("Dockerfile.api", "supperjumpin-api:dev"))
}

// Bot builds the Discord bot Docker image as supperjumpin-bot:dev.
func (Build) Bot() error {
	return runner.Run(buildImageCommand("Dockerfile.bot", "supperjumpin-bot:dev"))
}
