package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CommandSpec struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

type Runner interface {
	Run(CommandSpec) error
	Output(CommandSpec) (string, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(spec CommandSpec) error {
	cmd := exec.Command(spec.Name, spec.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if spec.Dir != "" {
		cmd.Dir = spec.Dir
	}
	if len(spec.Env) > 0 {
		cmd.Env = append(os.Environ(), spec.Env...)
	}
	return cmd.Run()
}

func (ExecRunner) Output(spec CommandSpec) (string, error) {
	cmd := exec.Command(spec.Name, spec.Args...)
	if spec.Dir != "" {
		cmd.Dir = spec.Dir
	}
	if len(spec.Env) > 0 {
		cmd.Env = append(os.Environ(), spec.Env...)
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

var runner Runner = ExecRunner{}

func runAll(specs ...CommandSpec) error {
	for _, spec := range specs {
		if err := runner.Run(spec); err != nil {
			return fmt.Errorf("run %s: %w", spec, err)
		}
	}
	return nil
}

func (c CommandSpec) String() string {
	parts := make([]string, 0, 1+len(c.Args))
	parts = append(parts, c.Name)
	parts = append(parts, c.Args...)
	return strings.Join(parts, " ")
}
