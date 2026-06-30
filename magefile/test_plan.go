package main

import (
	"fmt"
	"regexp"
	"strconv"
)

var totalCoveragePattern = regexp.MustCompile(`total:\s+\(statements\)\s+([\d.]+)%`)

func goTestCommand(moduleDir, coverProfile string, extraArgs []string) CommandSpec {
	return goTestCommandWithEnv(moduleDir, coverProfile, extraArgs, nil)
}

func goTestCommandWithEnv(moduleDir, coverProfile string, extraArgs []string, env []string) CommandSpec {
	args := []string{"test"}
	if coverProfile != "" {
		args = append(args, "-covermode=atomic", "-coverprofile="+coverProfile)
	}
	args = append(args, "./...")
	args = append(args, extraArgs...)
	return CommandSpec{Name: "go", Dir: moduleDir, Args: args, Env: env}
}

func goCoverFuncCommand(moduleDir, coverProfile string) CommandSpec {
	return CommandSpec{Name: "go", Dir: moduleDir, Args: []string{"tool", "cover", "-func=" + coverProfile}}
}

func parseTotalCoverage(output string) (float64, error) {
	match := totalCoveragePattern.FindStringSubmatch(output)
	if len(match) != 2 {
		return 0, fmt.Errorf("could not find total coverage in output")
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parse total coverage: %w", err)
	}
	return value, nil
}

func testDatabaseURLFromEnv(getenv func(string) string) string {
	if value := getenv("SUPPERJUMPIN_TEST_DATABASE_URL"); value != "" {
		return value
	}
	return DefaultTestDatabaseURL
}
