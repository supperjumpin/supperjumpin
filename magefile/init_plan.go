package main

import "fmt"

const (
	GoVersionRequirement       = ">=1.26.3 <1.27"
	DockerComposeRequirement   = ">=2.0.0"
	SQLCVersion                = "1.31.1"
	MigrateVersion             = "4.19.1"
	SQLCModulePath             = "github.com/sqlc-dev/sqlc/cmd/sqlc"
	MigrateModulePath          = "github.com/golang-migrate/migrate/v4/cmd/migrate"
	MigrateBuildTags           = "postgres"
	DefaultDocsPort            = "3456"
)

func goInstallCommand(binDir, modulePath, version string, extraArgs []string) CommandSpec {
	args := append([]string{"install"}, extraArgs...)
	args = append(args, fmt.Sprintf("%s@v%s", modulePath, version))
	return CommandSpec{
		Name: "go",
		Dir:  repoRoot(),
		Args: args,
		Env:  []string{"GOBIN=" + repoPath(binDir)},
	}
}
