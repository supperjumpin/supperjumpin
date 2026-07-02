# Build images only — no runtime deploy artifacts yet

The repo gains `Dockerfile.api` and `Dockerfile.bot`: multi-stage builds that produce a static Go binary in a pinned `golang:1.26` base, runnable as a container. `mage build:api` and `mage build:bot` invoke `docker build` against them. CI's `mage ci:build` job uses the same Dockerfiles.

These are **build images**, not **runtime deploy images**. There is no `docker-compose.yml` service for the API or the bot, no `EXPOSE` directive in production posture, no image registry push, no deploy target. The Dockerfiles exist to give CI a reproducible build environment (parity between local and GitHub Actions), to seed a future deployable image, and to make the build's Go version traceable to a single source.

The trade-off: we add two Dockerfiles and a small amount of CI surface now, but defer every hosted-infra concern. AGENTS.md's "Infrastructure is local-first: Docker Postgres, local dev bearer auth. Hosted infrastructure will be additive when introduced" still holds — these Dockerfiles are local-CI infra, not hosted infra.

## Why "build only" instead of "build + compose services"

- AGENTS.md is explicit that hosted infrastructure is "additive when introduced." Composing the API and bot into `docker-compose.yml` is a hosted-infra-shaped decision (it sets the deploy target, the env var surface, the port mapping) before there is a deploy to target.
- The current `docker-compose.yml` is Postgres-only and that single service does its job. Adding two more services to it is the kind of speculative infra the pivot has been deliberately avoiding.
- A multi-stage build-only Dockerfile is small (~30 lines), well-understood, and trivially upgradeable to a deployable image the moment a deploy target exists. Adding `compose` services now would mean removing them later (or carrying dead services in compose).

## Dockerfile shape

Both `Dockerfile.api` and `Dockerfile.bot` follow the same pattern:

- **Stage 1 (`build`):** `golang:1.26` base, copy `go.mod`/`go.sum`/`go.work`/`go.work.sum`, `go mod download`, copy source, `CGO_ENABLED=0 go build -o /out/<binary> ./apps/<app>/cmd/<binary>`.
- **Stage 2 (`runtime`):** `gcr.io/distroless/static-debian12:nonroot` base, copy the built binary from stage 1, set the non-root user. No shell, no package manager, no extras.

The distroless runtime image is small (~2MB), has no shell to attack, and runs as non-root by default. It is the right default for a Go binary with no CGO and no system library dependencies.

`mage build:api` and `mage build:bot` invoke `docker build -f Dockerfile.api -t supperjumpin-api:dev .` and the bot equivalent, from the repo root. The image tag is the dev tag — no registry push, no version tags, no multi-arch builds in this slice.

## Reintroduction

When a deploy target exists (Cloud Run, Fly.io, ECS, anything), the path is:
1. Add a multi-arch build step to `mage build:image:*` (e.g. `docker buildx build --platform linux/amd64,linux/arm64`).
2. Add a `mage push:image:*` target that tags for the registry and pushes.
3. Add the API and bot as services to `docker-compose.yml` (or its replacement) so local dev matches deploy.
4. Add the registry/deploy credentials to GitHub Actions secrets.

Each of those is a separate ADR-shaped decision when it lands.

## Status

accepted — adds build images for the API and bot; defers runtime/deploy infrastructure to a future slice.
