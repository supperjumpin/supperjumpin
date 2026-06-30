package main

import (
	"math"
	"reflect"
	"testing"
)

func TestGoTestCommand(t *testing.T) {
	t.Run("without coverage", func(t *testing.T) {
		got := goTestCommand(repoPath("apps", "bot-discord"), "", nil)
		want := CommandSpec{Name: "go", Dir: repoPath("apps", "bot-discord"), Args: []string{"test", "./..."}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("goTestCommand() = %#v, want %#v", got, want)
		}
	})

	t.Run("with coverage", func(t *testing.T) {
		got := goTestCommand(repoPath("apps", "api"), repoPath("coverage", "api.coverprofile"), []string{"-run", "TestSomething"})
		want := CommandSpec{
			Name: "go",
			Dir:  repoPath("apps", "api"),
			Args: []string{"test", "-covermode=atomic", "-coverprofile=" + repoPath("coverage", "api.coverprofile"), "./...", "-run", "TestSomething"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("goTestCommand() = %#v, want %#v", got, want)
		}
	})

	t.Run("with env", func(t *testing.T) {
		got := goTestCommandWithEnv(repoPath("apps", "api"), "", nil, []string{"SUPPERJUMPIN_TEST_DATABASE_URL=postgres://example"})
		want := CommandSpec{Name: "go", Dir: repoPath("apps", "api"), Args: []string{"test", "./..."}, Env: []string{"SUPPERJUMPIN_TEST_DATABASE_URL=postgres://example"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("goTestCommandWithEnv() = %#v, want %#v", got, want)
		}
	})
}

func TestParseTotalCoverage(t *testing.T) {
	got, err := parseTotalCoverage("github.com/example/pkg\tcoverage: 12.3% of statements\ntotal:\t(statements)\t78.9%\n")
	if err != nil {
		t.Fatalf("parseTotalCoverage returned error: %v", err)
	}
	if math.Abs(got-78.9) > 0.001 {
		t.Fatalf("parseTotalCoverage() = %v, want 78.9", got)
	}
}

func TestParseTotalCoverageErrorsWithoutTotalLine(t *testing.T) {
	if _, err := parseTotalCoverage("coverage: 12.3% of statements\n"); err == nil {
		t.Fatal("parseTotalCoverage() expected error for missing total line")
	}
}

func TestTestDatabaseURLFromEnv(t *testing.T) {
	got := testDatabaseURLFromEnv(func(key string) string {
		if key == "SUPPERJUMPIN_TEST_DATABASE_URL" {
			return "postgres://custom:custom@db:5432/custom_test?sslmode=disable"
		}
		return ""
	})
	if got != "postgres://custom:custom@db:5432/custom_test?sslmode=disable" {
		t.Fatalf("testDatabaseURLFromEnv() = %q", got)
	}

	fallback := testDatabaseURLFromEnv(func(string) string { return "" })
	if fallback != DefaultTestDatabaseURL {
		t.Fatalf("testDatabaseURLFromEnv() fallback = %q, want %q", fallback, DefaultTestDatabaseURL)
	}
}
