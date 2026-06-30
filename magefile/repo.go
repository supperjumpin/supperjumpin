package main

import (
	"path/filepath"
	"runtime"
)

func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Dir(filepath.Dir(file))
}

func repoPath(elem ...string) string {
	parts := append([]string{repoRoot()}, elem...)
	return filepath.Clean(filepath.Join(parts...))
}
