package main

import (
	"reflect"
	"testing"
)

func TestAPIDevCommand(t *testing.T) {
	got := apiDevCommand(func(key string) string {
		switch key {
		case "SUPPERJUMPIN_DATABASE_URL":
			return "postgres://custom:custom@db:5432/custom?sslmode=disable"
		case "SUPPERJUMPIN_DEV_AUTH_TOKEN":
			return "alice-token"
		case "SUPPERJUMPIN_ADAPTER_TOKEN":
			return "adapter-token"
		default:
			return ""
		}
	})
	want := CommandSpec{
		Name: "go",
		Dir:  repoPath("apps", "api"),
		Args: []string{"run", "./cmd/api"},
		Env: []string{
			"SUPPERJUMPIN_DATABASE_URL=postgres://custom:custom@db:5432/custom?sslmode=disable",
			"SUPPERJUMPIN_DEV_AUTH_TOKEN=alice-token",
			"SUPPERJUMPIN_ADAPTER_TOKEN=adapter-token",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("apiDevCommand() = %#v, want %#v", got, want)
	}
}

func TestBotDevCommandDefaults(t *testing.T) {
	got := botDevCommand(func(string) string { return "" })
	want := CommandSpec{
		Name: "go",
		Dir:  repoPath("apps", "bot-discord"),
		Args: []string{"run", "./cmd/bot"},
		Env: []string{
			"SUPPERJUMPIN_BOT_TOKEN=Bot dev-placeholder-token",
			"SUPPERJUMPIN_ADAPTER_TOKEN=dev-token",
			"SUPPERJUMPIN_API_BASE_URL=http://localhost:8080",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("botDevCommand() = %#v, want %#v", got, want)
	}
}

func TestBuildImageCommand(t *testing.T) {
	got := buildImageCommand("Dockerfile.api", "supperjumpin-api:dev")
	want := CommandSpec{Name: "docker", Dir: repoRoot(), Args: []string{"build", "-f", "Dockerfile.api", "-t", "supperjumpin-api:dev", "."}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildImageCommand() = %#v, want %#v", got, want)
	}
}

func TestSQLCGenerateCommand(t *testing.T) {
	got := sqlcGenerateCommand("bin")
	want := CommandSpec{Name: repoPath("bin", "sqlc"), Dir: repoPath("apps", "api"), Args: []string{"generate"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sqlcGenerateCommand() = %#v, want %#v", got, want)
	}
}

func TestGitDiffExitCodeCommand(t *testing.T) {
	got := gitDiffExitCodeCommand("apps/api/internal/db")
	want := CommandSpec{Name: "git", Dir: repoRoot(), Args: []string{"diff", "--exit-code", "--", "apps/api/internal/db"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("gitDiffExitCodeCommand() = %#v, want %#v", got, want)
	}
}

func TestGoVetCommand(t *testing.T) {
	got := goVetCommand(repoPath("apps", "api"))
	want := CommandSpec{Name: "go", Dir: repoPath("apps", "api"), Args: []string{"vet", "./..."}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("goVetCommand() = %#v, want %#v", got, want)
	}
}
