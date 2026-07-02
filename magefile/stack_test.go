package main

import "testing"

func TestStackProcessArtifacts(t *testing.T) {
	artifacts := stackProcessArtifacts("api")
	if artifacts.PIDPath != repoPath(".mage", "api.pid") {
		t.Fatalf("PIDPath = %q, want %q", artifacts.PIDPath, repoPath(".mage", "api.pid"))
	}
	if artifacts.LogPath != repoPath(".mage", "api.log") {
		t.Fatalf("LogPath = %q, want %q", artifacts.LogPath, repoPath(".mage", "api.log"))
	}
}
