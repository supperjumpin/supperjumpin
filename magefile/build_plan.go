package main

func buildImageCommand(dockerfile, tag string) CommandSpec {
	return CommandSpec{
		Name: "docker",
		Dir:  repoRoot(),
		Args: []string{"build", "-f", dockerfile, "-t", tag, "."},
	}
}

func sqlcGenerateCommand(binDir string) CommandSpec {
	return CommandSpec{
		Name: repoPath(binDir, "sqlc"),
		Dir:  repoPath("apps", "api"),
		Args: []string{"generate"},
	}
}

func gitDiffExitCodeCommand(paths ...string) CommandSpec {
	args := []string{"diff", "--exit-code"}
	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}
	return CommandSpec{Name: "git", Dir: repoRoot(), Args: args}
}
