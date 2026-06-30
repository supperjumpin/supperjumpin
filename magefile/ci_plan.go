package main

func goVetCommand(moduleDir string) CommandSpec {
	return CommandSpec{Name: "go", Dir: moduleDir, Args: []string{"vet", "./..."}}
}
