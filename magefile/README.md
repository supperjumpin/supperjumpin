# Mage Build Orchestration

`magefile/` is a small Go module that defines this repo's local development and CI commands. The public entry point is `mage -l` from the repository root.

## Structure

- `*_targets.go`: Mage targets and namespaces exposed to users.
- `*_plan.go`: pure helpers that return command plans.
- `command.go`: the execution adapter for command plans.
- `*_test.go`: tests for command planning and orchestration helpers.

## Common Commands

```sh
mage -l
mage init:all
mage db:reset
mage dev:api
mage test -coverage
mage ci:all
```

Keep target comments current so `mage -l` stays self-documenting.
