# Magefile Guide

This file extends the root `AGENTS.md` for the `magefile/` Go module. Keep this guide focused on build orchestration rules that are not already covered at the repo root.

## Purpose

- `magefile/` owns local development, test, CI, Docker image, docs, and generated-code orchestration.
- Mage targets are the public interface. `mage -l` must remain the primary discovery surface for humans and agents.
- Command-plan helpers build `CommandSpec` values so behavior can be tested without shelling out.

## Layout

- `*_targets.go` files contain Mage namespaces and target functions.
- `*_plan.go` files contain pure command-building helpers.
- `*_test.go` files test command plans and small orchestration helpers.
- `command.go` is the narrow execution adapter; avoid spreading direct `exec.Command` calls.

## Rules

- Add a one-sentence doc comment to every exported Mage target so it appears in `mage -l`.
- Keep targets thin: validate inputs, compose command plans, then run through `runner` or `runAll`.
- Prefer testable helpers that accept explicit inputs or injected environment readers.
- Do not add new external tools without pinning their versions in code or `go.mod` and documenting the user-facing target.
- Keep generated artifacts outside this module unless they are Mage-owned state such as `.mage/` process files.

## Verification

- Run `go test ./...` from `magefile/` after changing command plans or helper logic.
- Run `mage -l` from the repo root after changing target names or comments.
- Use the smallest target-specific command when verifying behavior that has side effects.
