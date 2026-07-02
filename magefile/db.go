package main

import (
	"fmt"
	"net/url"
	"strings"
)

// IsSafeToReset reports whether it is safe to destructively reset a database
// with the given name. A name is safe when it ends with "_test" or when the
// caller has explicitly opted in via allowUnsafe.
func IsSafeToReset(dbName string, allowUnsafe bool) bool {
	if allowUnsafe {
		return true
	}
	return strings.HasSuffix(dbName, "_test")
}

// ParseDatabaseName extracts the database name from a Postgres connection URL.
// Returns an error when the URL is malformed or has no database component.
func ParseDatabaseName(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse database url: %w", err)
	}
	name := strings.TrimPrefix(parsed.Path, "/")
	if name == "" {
		return "", fmt.Errorf("could not parse database name from url: %q", rawURL)
	}
	return name, nil
}

// BuildAdminURL rewrites a database URL to target the postgres admin database.
func BuildAdminURL(databaseURL string) (string, error) {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("parse database url: %w", err)
	}
	parsed.Path = "/postgres"
	return parsed.String(), nil
}

// DescribeDatabaseURL returns a host:port/path summary for logs.
func DescribeDatabaseURL(databaseURL string) (string, error) {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("parse database url: %w", err)
	}
	return fmt.Sprintf("%s%s%s", parsed.Hostname(), portSuffix(parsed.Port()), parsed.Path), nil
}

func DockerComposeArgsFor(action string) ([]string, error) {
	switch action {
	case "up":
		return []string{"compose", "up", "-d", "postgres"}, nil
	case "down":
		return []string{"compose", "stop", "postgres"}, nil
	default:
		return nil, fmt.Errorf("unsupported db lifecycle action: %s", action)
	}
}

func portSuffix(port string) string {
	if port == "" {
		return ""
	}
	return ":" + port
}
