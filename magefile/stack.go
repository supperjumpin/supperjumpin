package main

import (
	"errors"
	"os"
	"strconv"
	"syscall"
)

type processArtifacts struct {
	PIDPath string
	LogPath string
}

func stackProcessArtifacts(name string) processArtifacts {
	return processArtifacts{
		PIDPath: repoPath(".mage", name+".pid"),
		LogPath: repoPath(".mage", name+".log"),
	}
}

func stopProcessArtifacts(artifacts processArtifacts, signal func(int, os.Signal) error) error {
	data, err := os.ReadFile(artifacts.PIDPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}
	if err := signal(pid, syscall.SIGTERM); err != nil && !isAlreadyStoppedProcess(err) {
		return err
	}
	return os.Remove(artifacts.PIDPath)
}

func signalProcess(pid int, signal os.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(signal)
}

func isAlreadyStoppedProcess(err error) bool {
	return errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH)
}
