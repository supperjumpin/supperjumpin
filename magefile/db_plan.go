package main

import (
	"fmt"
	"time"
)

const (
	DefaultDevelopmentDatabaseURL = "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable"
	DefaultTestDatabaseURL        = "postgres://postgres:postgres@localhost:5432/supperjumpin_test?sslmode=disable"
	defaultWaitTimeout            = 30 * time.Second
	defaultWaitInterval           = 500 * time.Millisecond
)

func dbLifecycleCommand(action string) (CommandSpec, error) {
	args, err := DockerComposeArgsFor(action)
	if err != nil {
		return CommandSpec{}, err
	}
	return CommandSpec{Name: "docker", Args: args}, nil
}

func dbMigrateCommand(databaseURL, binDir string) CommandSpec {
	return CommandSpec{
		Name: repoPath(binDir, "migrate"),
		Args: []string{"-database", databaseURL, "-path", repoPath("apps", "api", "db", "migrations"), "up"},
	}
}

func dbResetCommands(databaseURL string, isLocalDocker bool, binDir string) ([]CommandSpec, error) {
	dbName, err := ParseDatabaseName(databaseURL)
	if err != nil {
		return nil, err
	}
	adminURL, err := BuildAdminURL(databaseURL)
	if err != nil {
		return nil, err
	}
	return []CommandSpec{
		psqlCommand(adminURL, fmt.Sprintf("DROP DATABASE IF EXISTS %q;", dbName), isLocalDocker),
		psqlCommand(adminURL, fmt.Sprintf("CREATE DATABASE %q;", dbName), isLocalDocker),
		dbMigrateCommand(databaseURL, binDir),
	}, nil
}

func psqlCommand(databaseURL, sql string, isLocalDocker bool) CommandSpec {
	if isLocalDocker {
		return CommandSpec{Name: "docker", Args: []string{"compose", "exec", "-T", "postgres", "psql", databaseURL, "-c", sql}}
	}
	return CommandSpec{Name: "psql", Args: []string{databaseURL, "-c", sql}}
}

func postgresReadyProbe() CommandSpec {
	return CommandSpec{Name: "docker", Args: []string{"compose", "exec", "-T", "postgres", "pg_isready", "-h", "localhost", "-p", "5432", "-U", "postgres"}}
}

func waitForPostgresReady(timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := runner.Run(postgresReadyProbe()); err == nil {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("postgres did not become ready in %s", timeout)
}

func migrateBinDirFromEnv(getenv func(string) string) string {
	if dir := getenv("SUPPERJUMPIN_MIGRATE_BIN_DIR"); dir != "" {
		return dir
	}
	return "bin"
}
