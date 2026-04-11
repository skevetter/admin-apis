# admin-apis Tooling Upgrade

**Date:** 2026-04-11
**Status:** Draft
**Repo:** admin-apis (Go library, no binary artifacts)

## Goal

Bring admin-apis tooling in line with the devpod ecosystem (provider-aws, devpod) for consistency in linting, formatting, CI, dependency management, and release automation.

## Current State

- **Justfile** with lint, gen, check-structalign tasks
- **GitHub Actions:** go.yml (build + go generate check), lint.yaml (golangci-lint), notify_repositories.yaml (dispatch to downstream repos on definitions/ changes)
- **No:** pre-commit, Taskfile, .golangci.yaml config, goreleaser, renovate, release automation, conventional commits enforcement
- **Outdated:** action versions (checkout@v4, setup-go@v4/v5, golangci-lint-action@v4), Go dependencies

## Design

Each change ships as a separate PR, merged in order. Dependencies between PRs are noted.

### PR 0: Module Path Migration & Devendor

Change Go module path from `github.com/loft-sh/admin-apis` to `github.com/skevetter/admin-apis`. Update all imports, generated files, and code generation commands. Remove vendor directory if present.

### PR 1: Pre-commit

Add `.pre-commit-config.yaml` matching the standard devpod config:

- pre-commit-hooks v6.0.0 (trailing-whitespace, end-of-file-fixer, check-yaml, check-added-large-files, check-merge-conflict)
- commitlint v9.24.0 (conventional commit enforcement)
- shellcheck-py v0.11.0.1
- shfmt v3.12.0-2 (4-space indent)
- actionlint v1.7.12
- gitleaks v8.30.1
- commitizen v4.13.9
- golangci-lint v2.11.4 (lint + fmt)

Add `commitlint.config.mjs` extending `@commitlint/config-conventional`.

Add `requirements.txt` for pre-commit Python dependency (needed for CI).

**Lint fix pass:** Run pre-commit on all files and fix any violations (trailing whitespace, EOF, yaml, shell scripts, etc.).

### PR 2: Golangci-lint Config

Add `.golangci.yaml` with the standard devpod config:

- **Run:** 5m timeout, 8 concurrent
- **Formatters:** gci, gofumpt, goimports, golines
- **Linters (23):** cyclop, decorder, dupl, errcheck, fatcontext, forbidigo, funcorder, funlen, goconst, gocritic, godot, gosec, lll, misspell, modernize, nestif, revive, staticcheck, unparam, unused, whitespace
- **Settings:** cyclop max-complexity=8, revive argument-limit=4, function-result-limit=3

**Lint fix pass:** Run golangci-lint and fix all violations. This may be the largest PR due to formatter/linter changes across the codebase.

### PR 3: Taskfile

Replace `Justfile` with `Taskfile.yml` carrying over existing tasks:

- `task lint` — run golangci-lint
- `task gen` — run code generation (deepcopy-gen, openapi-gen, go generate)
- `task check-structalign` — run betteralign

Remove the `Justfile`.

### PR 4: Dependency Updates

- `go get -u ./...` and `go mod tidy` to update Go module dependencies
- Update GitHub Actions to latest versions across all workflows:
  - actions/checkout@v4 -> @v6
  - actions/setup-go@v4/@v5 -> @v6
  - golangci/golangci-lint-action@v4 -> @v9

### PR 5: Renovate

Add `renovate.json`:

```json
{
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "automerge": false,
    "automergeType": "pr",
    "automergeStrategy": "squash",
    "extends": ["config:recommended"],
    "pre-commit": { "enabled": true },
    "dependencyDashboard": false,
    "prConcurrentLimit": 1,
    "platformAutomerge": true,
    "postUpdateOptions": ["gomodTidy"],
    "packageRules": [
        { "matchUpdateTypes": ["major"], "automerge": false, "prPriority": -1 },
        { "matchUpdateTypes": ["minor", "patch"], "prPriority": 10 },
        { "matchManagers": ["github-actions"], "automerge": true, "prPriority": 5 },
        { "matchManagers": ["pre-commit"], "automerge": true, "prPriority": 4 },
        { "matchManagers": ["gomod"], "automerge": true, "prPriority": 3 }
    ]
}
```

Remove `notify_repositories.yaml` workflow — renovate handles downstream dependency updates.

### PR 6: Release Automation (release-please)

Add `.github/workflows/release.yml`:

- Triggered on push to main
- Uses `googleapis/release-please-action`
- Configured for Go (`release-type: go`)
- Creates/updates a release PR with changelog from conventional commits
- Merging the release PR creates a GitHub Release with semver tag (e.g., v1.0.0)

Add `.release-please-manifest.json` and `release-please-config.json` for configuration.

### PR 7: CI Modernization

Add/update GitHub Actions workflows to match devpod standard:

- **commit.yml** — Check signed commits (1Password) + commitlint on PRs
- **pre-commit.yml** — Run pre-commit on all files for push to main and PRs
- **lint.yml** — Updated golangci-lint workflow with latest action versions
- **workflow-approval.yml** — Auto-approve workflows for PRs

Remove the existing `go.yml` workflow. The `go generate` freshness check moves to a Taskfile task (`task gen:check`) and is run in the lint or pre-commit CI workflow.

## Out of Scope

- **Goreleaser** — admin-apis is a library with no binaries to build
- **Biome/JS formatting** — no JavaScript/TypeScript in this repo
- **Megalinter** — not part of the standard devpod tooling

## Risks and Notes

- **PR 2 (golangci-lint)** will likely be the largest PR due to formatter and linter fixes across the codebase. Consider auto-fixing what's possible and reviewing manually.
- **PR 4 (dependency updates)** may introduce breaking changes from major version bumps. Review changelogs for major updates.
- **PR 5 (renovate)** requires the Renovate GitHub App to be installed on the repo.
- **PR 6 (release-please)** — repo has no existing version tags, so no bootstrap needed. First release will be v1.0.0 (or as configured).
