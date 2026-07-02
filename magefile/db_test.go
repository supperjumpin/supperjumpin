package main

import (
	"reflect"
	"testing"
)

func TestIsSafeToReset(t *testing.T) {
	tests := []struct {
		name        string
		dbName      string
		allowUnsafe bool
		want        bool
	}{
		{name: "dev database is unsafe", dbName: "supperjumpin", allowUnsafe: false, want: false},
		{name: "test database is safe", dbName: "supperjumpin_test", allowUnsafe: false, want: true},
		{name: "custom test database is safe", dbName: "myapp_test", allowUnsafe: false, want: true},
		{name: "allowUnsafe overrides dev database", dbName: "supperjumpin", allowUnsafe: true, want: true},
		{name: "allowUnsafe is redundant for test database", dbName: "supperjumpin_test", allowUnsafe: true, want: true},
		{name: "empty name is unsafe", dbName: "", allowUnsafe: false, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafeToReset(tt.dbName, tt.allowUnsafe); got != tt.want {
				t.Errorf("IsSafeToReset(%q, %v) = %v, want %v", tt.dbName, tt.allowUnsafe, got, tt.want)
			}
		})
	}
}

func TestParseDatabaseName(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "standard postgres URL",
			url:  "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable",
			want: "supperjumpin",
		},
		{
			name: "test database",
			url:  "postgres://postgres:postgres@localhost:5432/supperjumpin_test?sslmode=disable",
			want: "supperjumpin_test",
		},
		{
			name: "URL with no query string",
			url:  "postgres://user:pass@host:5432/mydb",
			want: "mydb",
		},
		{
			name:    "URL with no database path",
			url:     "postgres://user:pass@host:5432/",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDatabaseName(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDatabaseName(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDatabaseName(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestBuildAdminURL(t *testing.T) {
	got, err := BuildAdminURL("postgres://postgres:postgres@localhost:5432/supperjumpin_test?sslmode=disable")
	if err != nil {
		t.Fatalf("BuildAdminURL returned error: %v", err)
	}
	want := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	if got != want {
		t.Fatalf("BuildAdminURL() = %q, want %q", got, want)
	}
}

func TestDescribeDatabaseURL(t *testing.T) {
	got, err := DescribeDatabaseURL("postgres://postgres:postgres@db.internal:5432/supperjumpin?sslmode=disable")
	if err != nil {
		t.Fatalf("DescribeDatabaseURL returned error: %v", err)
	}
	want := "db.internal:5432/supperjumpin"
	if got != want {
		t.Fatalf("DescribeDatabaseURL() = %q, want %q", got, want)
	}
}

func TestDockerComposeArgsFor(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		want    []string
		wantErr bool
	}{
		{name: "up mode starts local postgres in detached mode", action: "up", want: []string{"compose", "up", "-d", "postgres"}},
		{name: "down mode stops local postgres without deleting data", action: "down", want: []string{"compose", "stop", "postgres"}},
		{name: "unknown action errors", action: "bounce", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DockerComposeArgsFor(tt.action)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DockerComposeArgsFor(%q) error = %v, wantErr %v", tt.action, err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("DockerComposeArgsFor(%q) = %#v, want %#v", tt.action, got, tt.want)
			}
		})
	}
}

func TestPsqlCommand(t *testing.T) {
	t.Run("local docker uses docker compose exec", func(t *testing.T) {
		got := psqlCommand("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "DROP DATABASE IF EXISTS \"supperjumpin\";", true)
		want := CommandSpec{
			Name: "docker",
			Args: []string{"compose", "exec", "-T", "postgres", "psql", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "-c", "DROP DATABASE IF EXISTS \"supperjumpin\";"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("psqlCommand() = %#v, want %#v", got, want)
		}
	})

	t.Run("remote database uses direct psql", func(t *testing.T) {
		got := psqlCommand("postgres://user:pass@db.internal:5432/postgres?sslmode=require", "SELECT 1;", false)
		want := CommandSpec{Name: "psql", Args: []string{"postgres://user:pass@db.internal:5432/postgres?sslmode=require", "-c", "SELECT 1;"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("psqlCommand() = %#v, want %#v", got, want)
		}
	})
}

func TestPostgresReadyProbeChecksTCPListener(t *testing.T) {
	got := postgresReadyProbe()
	want := CommandSpec{Name: "docker", Args: []string{"compose", "exec", "-T", "postgres", "pg_isready", "-h", "localhost", "-p", "5432", "-U", "postgres"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("postgresReadyProbe() = %#v, want %#v", got, want)
	}
}

func TestDBMigrateCommand(t *testing.T) {
	got := dbMigrateCommand("postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable", "bin")
	want := CommandSpec{
		Name: repoPath("bin", "migrate"),
		Args: []string{"-database", "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable", "-path", repoPath("apps", "api", "db", "migrations"), "up"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dbMigrateCommand() = %#v, want %#v", got, want)
	}
}

func TestDBResetCommands(t *testing.T) {
	got, err := dbResetCommands("postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable", true, "bin")
	if err != nil {
		t.Fatalf("dbResetCommands returned error: %v", err)
	}
	want := []CommandSpec{
		{Name: "docker", Args: []string{"compose", "exec", "-T", "postgres", "psql", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "-c", "DROP DATABASE IF EXISTS \"supperjumpin\";"}},
		{Name: "docker", Args: []string{"compose", "exec", "-T", "postgres", "psql", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "-c", "CREATE DATABASE \"supperjumpin\";"}},
		{Name: repoPath("bin", "migrate"), Args: []string{"-database", "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable", "-path", repoPath("apps", "api", "db", "migrations"), "up"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dbResetCommands() = %#v, want %#v", got, want)
	}
}
