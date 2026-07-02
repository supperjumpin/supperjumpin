//go:build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
)

type Init mg.Namespace

// Tools installs pinned local helper binaries into ./bin.
func (Init) Tools() error {
	fmt.Println("Checking local tool binaries in ./bin")
	if err := os.MkdirAll(repoPath("bin"), 0o755); err != nil {
		return err
	}
	fmt.Printf("Installing sqlc v%s into ./bin\n", SQLCVersion)
	if err := runner.Run(goInstallCommand("bin", SQLCModulePath, SQLCVersion, nil)); err != nil {
		return err
	}
	fmt.Printf("Installing migrate v%s (%s) into ./bin\n", MigrateVersion, MigrateBuildTags)
	if err := runner.Run(goInstallCommand("bin", MigrateModulePath, MigrateVersion, []string{"-tags", MigrateBuildTags})); err != nil {
		return err
	}
	fmt.Println("Local tools are ready")
	return nil
}

// Host prints required host tool versions.
func (Init) Host() error {
	fmt.Printf("Go requirement: %s\n", GoVersionRequirement)
	fmt.Printf("Docker Compose requirement: %s\n", DockerComposeRequirement)
	return nil
}

// All installs local tools and prints common next commands.
func (Init) All() error {
	mg.SerialDeps(Init{}.Host, Init{}.Tools)
	fmt.Println("\nSetup complete. Useful commands:")
	fmt.Println("  mage db:reset")
	fmt.Println("  mage dev:api")
	fmt.Println("  mage dev:bot")
	fmt.Println("  mage test -coverage")
	fmt.Println("  mage docs")
	return nil
}
