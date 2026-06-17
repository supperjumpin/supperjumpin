package main

import "testing"

func TestRequiredDatabaseURLReturnsSUPPERJUMPINDatabaseURL(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_DATABASE_URL", "postgres://user:pass@primary:5432/supperjumpin?sslmode=disable")
	t.Setenv("DATABASE_URL", "postgres://user:pass@ambient:5432/supperjumpin?sslmode=disable")

	got, err := requiredDatabaseURL()
	if err != nil {
		t.Fatalf("expected database URL, got error: %v", err)
	}

	want := "postgres://user:pass@primary:5432/supperjumpin?sslmode=disable"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRequiredDatabaseURLFailsWithoutSUPPERJUMPINDatabaseURL(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_DATABASE_URL", "")
	t.Setenv("DATABASE_URL", "postgres://user:pass@ambient:5432/supperjumpin?sslmode=disable")

	got, err := requiredDatabaseURL()
	if err == nil {
		t.Fatalf("expected error, got database URL %q", got)
	}

	want := "SUPPERJUMPIN_DATABASE_URL is required for durable Supperjumpin API state"
	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}
