//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Up starts Postgres, API, and bot as a local background stack.
func Up() error {
	db := DB{}
	if err := db.Up(); err != nil {
		return err
	}
	if err := startBackgroundProcess("api", apiDevCommand(os.Getenv)); err != nil {
		return err
	}
	if err := startBackgroundProcess("bot", botDevCommand(os.Getenv)); err != nil {
		return err
	}
	fmt.Printf("Stack started. Logs: %s and %s\n", stackProcessArtifacts("api").LogPath, stackProcessArtifacts("bot").LogPath)
	return nil
}

// Down stops the local background stack.
func Down() error {
	if err := stopBackgroundProcess("bot"); err != nil {
		return err
	}
	if err := stopBackgroundProcess("api"); err != nil {
		return err
	}
	db := DB{}
	return db.Down()
}

func startBackgroundProcess(name string, spec CommandSpec) error {
	artifacts := stackProcessArtifacts(name)
	if err := os.MkdirAll(repoPath(".mage"), 0o755); err != nil {
		return err
	}
	logFile, err := os.OpenFile(artifacts.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	cmd := exec.Command(spec.Name, spec.Args...)
	cmd.Dir = spec.Dir
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(), spec.Env...)
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return err
	}
	if err := os.WriteFile(artifacts.PIDPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0o644); err != nil {
		_ = cmd.Process.Kill()
		logFile.Close()
		return err
	}
	return logFile.Close()
}

func stopBackgroundProcess(name string) error {
	return stopProcessArtifacts(stackProcessArtifacts(name), signalProcess)
}
