# Admin-APIs Tooling Modernization Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring admin-apis in line with the provider repos (AWS, gcloud) for linting, formatting, pre-commit hooks, CI, task automation, dependency management, and release automation.

**Architecture:** Six sequential PRs, each independently mergeable. Each PR adds one tooling concern. Order chosen so later PRs don't need to fix earlier ones (e.g., golangci config before pre-commit, since pre-commit uses golangci-lint). Admin-apis is a **library** (no binary), so goreleaser is replaced with a lightweight tag+changelog release workflow.

**Tech Stack:** golangci-lint v2, pre-commit, Taskfile v3, Renovate, GitHub Actions

**Key difference from providers:** No goreleaser (no binary artifacts). Release automation creates GitHub Releases with auto-generated changelogs from conventional commits when a tag is pushed. The existing `notify_repositories` workflow already handles downstream signaling on `main` push.

**Reference repos:**
- `~/ws/devpod/provider-aws` (primary reference)
- `~/ws/devpod/devpod-provider-gcloud` (secondary reference)

**Related issues:**
- https://github.com/skevetter/devpod/issues/702 (DigitalOcean provider needs releases)
- https://github.com/skevetter/devpod/issues/606 (set up all providers uniformly)

---

## PR 1: Golangci-lint Configuration

**Branch:** `tooling/golangci-lint`
**Base:** `main`

Adds `.golangci.yaml` (v2 format) matching provider standard. Updates the existing `lint.yaml` workflow to use current action versions and `go-version-file`.

### Task 1.1: Add .golangci.yaml

**Files:**
- Create: `.golangci.yaml`

- [ ] **Step 1: Create `.golangci.yaml`**

```yaml
version: "2"
run:
  timeout: 5m
  concurrency: 8
formatters:
  enable:
    - gci
    - gofumpt
    - goimports
    - golines
linters:
  enable:
    - cyclop
    - decorder
    - dupl
    - errcheck
    - fatcontext
    - forbidigo
    - funcorder
    - funlen
    - goconst
    - gocritic
    - godot
    - gosec
    - lll
    - misspell
    - modernize
    - nestif
    - revive
    - staticcheck
    - unparam
    - unused
    - whitespace
  settings:
    cyclop:
      max-complexity: 8
    revive:
      rules:
        - name: argument-limit
          arguments: [4]
        - name: function-result-limit
          arguments: [3]
```

- [ ] **Step 2: Run linter locally to check for issues**

Run: `golangci-lint run --timeout=5m`

Note: There will likely be many findings on first run. This is expected — the goal of this PR is to add the config, not fix all lint issues. Use `//nolint` sparingly or adjust settings if the volume is overwhelming. Consider adding `issues.max-issues-per-linter: 0` and `issues.max-same-issues: 0` if you want to see everything, or start with `only-new-issues: true` in CI.

- [ ] **Step 3: Commit**

```bash
git add .golangci.yaml
git commit -m "feat: add golangci-lint v2 configuration

Standardize linting and formatting configuration to match provider
repositories. Enables 21 linters and 4 formatters."
```

### Task 1.2: Update lint workflow

**Files:**
- Modify: `.github/workflows/lint.yaml`

- [ ] **Step 1: Replace `.github/workflows/lint.yaml`**

```yaml
name: Lint

on:
  push:
    branches: [main]
  pull_request:
  workflow_dispatch:

concurrency:
  group: ${{ github.workflow }}-${{ github.event.pull_request.number || github.ref }}
  cancel-in-progress: true

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - name: setup Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v9
        with:
          version: latest
          only-new-issues: true
```

Changes from current:
- `actions/checkout` v4 → v6
- `actions/setup-go` v5 → v6, using `go-version-file` instead of hardcoded version
- `golangci-lint-action` v4 → v9
- Removed `release` trigger (linting on release creation is unnecessary)
- Added `push: branches: [main]` and `workflow_dispatch` triggers
- Added `only-new-issues: true` so existing code doesn't block PRs
- Removed `cache: false` (let the action manage its cache)

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/lint.yaml
git commit -m "feat: modernize lint workflow

Update action versions, use go-version-file, add only-new-issues
to avoid blocking PRs on pre-existing lint findings."
```

- [ ] **Step 3: Push and open PR**

```bash
git push -u origin tooling/golangci-lint
gh pr create --title "feat: add golangci-lint v2 configuration" --body "$(cat <<'EOF'
## Summary
- Adds `.golangci.yaml` with 21 linters and 4 formatters matching provider repos
- Modernizes lint CI workflow (action versions v4→v6/v9, `go-version-file`, `only-new-issues`)

## Test plan
- [ ] Lint workflow passes on this PR
- [ ] `golangci-lint run` works locally with the new config
EOF
)"
```

---

## PR 2: Pre-commit Hooks

**Branch:** `tooling/pre-commit`
**Base:** `main` (or stack on PR 1 if using stacked branches)

Adds `.pre-commit-config.yaml` and CI workflow. Adapts the provider config for a library repo (no shell scripts to lint with shellcheck/shfmt — include them anyway for future-proofing, they're no-ops without matching files).

### Task 2.1: Add pre-commit config and requirements

**Files:**
- Create: `.pre-commit-config.yaml`
- Create: `requirements.txt`

- [ ] **Step 1: Create `requirements.txt`**

```
pre_commit>=4.3.0
```

- [ ] **Step 2: Create `.pre-commit-config.yaml`**

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v6.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
        args: ["--allow-multiple-documents"]
      - id: check-added-large-files
      - id: check-merge-conflict
  - repo: https://github.com/alessandrojcm/commitlint-pre-commit-hook
    rev: v9.24.0
    hooks:
      - id: commitlint
        stages: [commit-msg]
        language_version: lts
        additional_dependencies:
          - "@commitlint/config-conventional"
        verbose: true
  - repo: https://github.com/shellcheck-py/shellcheck-py
    rev: v0.11.0.1
    hooks:
      - id: shellcheck
        args:
          [
            "-e",
            "SC1091,SC1103,SC2148,SC2034,SC1090,SC1009,SC1054,SC1056,SC1072,SC1073,SC1083,SC2329",
          ]
  - repo: https://github.com/scop/pre-commit-shfmt
    rev: v3.12.0-2
    hooks:
      - id: shfmt
        args: ["-i", "4", "-ci", "-w"]
  - repo: https://github.com/rhysd/actionlint
    rev: v1.7.12
    hooks:
      - id: actionlint
        files: '^\.github/workflows/'
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.30.1
    hooks:
      - id: gitleaks
  - repo: https://github.com/commitizen-tools/commitizen
    rev: v4.13.9
    hooks:
      - id: commitizen
        stages: [commit-msg]
  - repo: https://github.com/golangci/golangci-lint
    rev: v2.11.4
    hooks:
      - id: golangci-lint
      - id: golangci-lint-fmt
```

- [ ] **Step 3: Install hooks locally and test**

Run:
```bash
pip install pre-commit
pre-commit install
pre-commit install --hook-type commit-msg
pre-commit run --all-files
```

Note: First run will likely show formatting fixes. Stage and include them in the commit.

- [ ] **Step 4: Commit**

```bash
git add .pre-commit-config.yaml requirements.txt
# If pre-commit made formatting changes, add those too:
git add -u
git commit -m "feat: add pre-commit hooks configuration

Adds 8 hook providers: file hygiene, commitlint, shellcheck, shfmt,
actionlint, gitleaks, commitizen, and golangci-lint."
```

### Task 2.2: Add pre-commit CI workflow

**Files:**
- Create: `.github/workflows/pre-commit.yml`

- [ ] **Step 1: Create `.github/workflows/pre-commit.yml`**

```yaml
name: Pre-commit

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  check:
    name: Pre-commit Checks
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-python@v6
        with:
          cache: pip
          python-version: 3.x
      - run: pip install -r requirements.txt
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
      - uses: golangci/golangci-lint-action@v9
        with:
          version: latest
          install-only: true
      - run: go install golang.org/x/tools/cmd/goimports@latest
      - uses: actions/cache@v5
        with:
          path: ~/.cache/pre-commit
          key: pre-commit-${{ hashFiles('.pre-commit-config.yaml') }}
      - run: pre-commit run --all-files
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/pre-commit.yml
git commit -m "feat: add pre-commit CI workflow

Runs all pre-commit hooks in CI on PRs and pushes to main."
```

### Task 2.3: Add commit check workflow

**Files:**
- Create: `.github/workflows/commit.yml`

- [ ] **Step 1: Create `.github/workflows/commit.yml`**

```yaml
name: Commit

on:
  pull_request:

jobs:
  commits:
    name: Check Commits
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    steps:
      - uses: actions/checkout@v6
      - uses: 1Password/check-signed-commits-action@v1
      - uses: wagoid/commitlint-github-action@v6
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/commit.yml
git commit -m "feat: add commit message and signing checks

Enforces conventional commits and GPG-signed commits on PRs."
```

- [ ] **Step 3: Push and open PR**

```bash
git push -u origin tooling/pre-commit
gh pr create --title "feat: add pre-commit hooks and CI" --body "$(cat <<'EOF'
## Summary
- Adds `.pre-commit-config.yaml` with 8 hook providers (matching provider repos)
- Adds pre-commit CI workflow
- Adds commit message and signing check workflow

## Test plan
- [ ] Pre-commit hooks run locally with `pre-commit run --all-files`
- [ ] Pre-commit CI workflow passes
- [ ] Commit check workflow passes
EOF
)"
```

---

## PR 3: Taskfile (Replace Justfile)

**Branch:** `tooling/taskfile`
**Base:** `main`

Replaces the Justfile with a Taskfile.yml, preserving all existing commands (`lint`, `gen`, `check-structalign`) and adding a standard structure.

### Task 3.1: Create Taskfile and remove Justfile

**Files:**
- Create: `Taskfile.yml`
- Delete: `Justfile`

- [ ] **Step 1: Create `Taskfile.yml`**

```yaml
version: "3"

tasks:
  lint:
    desc: Run golangci-lint for all packages
    cmd: golangci-lint run {{ .CLI_ARGS }}

  fmt:
    desc: Run golangci-lint formatters
    cmd: golangci-lint fmt {{ .CLI_ARGS }}

  gen:
    desc: Generate all Go related APIs and files
    cmds:
      - >-
        go run k8s.io/code-generator/cmd/deepcopy-gen@v0.28.1
        --go-header-file ./hack/boilerplate.go.txt
        --input-dirs ./pkg/licenseapi
        -O zz_generated.deepcopy
      - go generate ./...
      - >-
        go run k8s.io/kube-openapi/cmd/openapi-gen@v0.0.0-20260127142750-a19766b6e2d4
        --go-header-file ./hack/boilerplate.go.txt
        --output-pkg github.com/skevetter/admin-apis/pkg/licenseapi
        --output-dir ./pkg/licenseapi
        --output-file zz_generated.openapi.go
        --output-model-name-file zz_generated.model_name.go
        github.com/skevetter/admin-apis/pkg/licenseapi

  check:gen:
    desc: Verify generated code is up to date
    cmds:
      - task: gen
      - cmd: |
          if [ -n "$(git diff)" ]; then
            echo "go generate resulted in changed files. Run 'task gen' and commit the changes."
            exit 1
          fi

  check:structalign:
    desc: Check struct memory alignment and print potential improvements
    cmd: go run github.com/dkorunic/betteralign/cmd/betteralign@latest {{ .CLI_ARGS }} ./...
```

- [ ] **Step 2: Verify tasks work**

Run:
```bash
task lint
task gen
task check:structalign
```

- [ ] **Step 3: Delete Justfile**

```bash
rm Justfile
```

- [ ] **Step 4: Commit**

```bash
git add Taskfile.yml
git rm Justfile
git commit -m "feat: replace Justfile with Taskfile

Migrates all commands (lint, gen, check-structalign) to Taskfile v3
format. Adds fmt and check:gen tasks."
```

- [ ] **Step 5: Push and open PR**

```bash
git push -u origin tooling/taskfile
gh pr create --title "feat: replace Justfile with Taskfile" --body "$(cat <<'EOF'
## Summary
- Replaces Justfile with Taskfile.yml (standardizes on Taskfile across all repos)
- Preserves all existing commands: `lint`, `gen`, `check-structalign`
- Adds `fmt` (golangci-lint formatters) and `check:gen` (verify generated code)

## Test plan
- [ ] `task lint` works
- [ ] `task gen` works
- [ ] `task fmt` works
- [ ] `task check:gen` works
- [ ] `task check:structalign` works
EOF
)"
```

---

## PR 4: Modernize CI Workflows

**Branch:** `tooling/ci-modernization`
**Base:** `main`

Updates `go.yml` to use current action versions and `go-version-file`. Adds workflow-approval for Renovate/bot PRs.

### Task 4.1: Update go.yml workflow

**Files:**
- Modify: `.github/workflows/go.yml`

- [ ] **Step 1: Replace `.github/workflows/go.yml`**

```yaml
name: Go

on:
  push:
    branches: [main]
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod

      - name: Check go generate changes
        run: |
          go generate ./...
          if [[ ! -z $(git diff) ]]; then
            echo "go generate resulted in changed files. Run go generate and commit the changes."
            exit 1;
          fi
```

Changes:
- `actions/checkout` v4 → v6
- `actions/setup-go` v4 → v6
- `go-version: "1.22"` → `go-version-file: go.mod` (always uses the version from go.mod)
- Removed `branches: ["main"]` filter on `pull_request` (run on all PR branches)

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/go.yml
git commit -m "feat: modernize Go CI workflow

Update action versions, use go-version-file instead of hardcoded version."
```

### Task 4.2: Add workflow-approval

**Files:**
- Create: `.github/workflows/workflow-approval.yml`

- [ ] **Step 1: Create `.github/workflows/workflow-approval.yml`**

```yaml
name: Automatic Approve Workflow

on:
  workflow_dispatch:
  pull_request_target:
    types: [opened, synchronize, reopened]

jobs:
  automatic-approve:
    name: Approve Workflows
    runs-on: ubuntu-latest
    steps:
      - uses: mheap/automatic-approve-action@v1
        with:
          token: ${{ secrets.GH_ACCESS_TOKEN }}
          workflows: "commit.yml,lint.yml,pre-commit.yml"
```

Note: Requires `GH_ACCESS_TOKEN` secret to be configured in the repo settings (same PAT used in provider repos).

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/workflow-approval.yml
git commit -m "feat: add automatic workflow approval for bot PRs

Enables Renovate and other bot PRs to have their CI workflows
automatically approved."
```

- [ ] **Step 3: Push and open PR**

```bash
git push -u origin tooling/ci-modernization
gh pr create --title "feat: modernize CI workflows" --body "$(cat <<'EOF'
## Summary
- Updates `go.yml` action versions (v4→v6) and uses `go-version-file`
- Adds workflow-approval for bot PRs (Renovate)

## Prerequisites
- `GH_ACCESS_TOKEN` secret must be configured in repo settings

## Test plan
- [ ] Go workflow passes on this PR
- [ ] Workflow approval triggers on PR events
EOF
)"
```

---

## PR 5: Renovate Configuration

**Branch:** `tooling/renovate`
**Base:** `main`

Adds Renovate for automated dependency updates with the same policy as provider repos.

### Task 5.1: Add renovate.json

**Files:**
- Create: `renovate.json`

- [ ] **Step 1: Create `renovate.json`**

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

Note: Requires the Renovate GitHub App to be installed on the `skevetter/admin-apis` repo. If not already installed, go to https://github.com/apps/renovate and enable it for this repo.

- [ ] **Step 2: Commit**

```bash
git add renovate.json
git commit -m "feat: add Renovate dependency management

Enables automated dependency updates for Go modules, GitHub Actions,
and pre-commit hooks. Auto-merges minor/patch, manual review for major."
```

- [ ] **Step 3: Push and open PR**

```bash
git push -u origin tooling/renovate
gh pr create --title "feat: add Renovate dependency management" --body "$(cat <<'EOF'
## Summary
- Adds `renovate.json` matching provider repo configuration
- Auto-merges minor/patch updates for gomod, GitHub Actions, and pre-commit
- Major updates require manual review
- Serialized PRs (`prConcurrentLimit: 1`)

## Prerequisites
- Renovate GitHub App must be installed on this repo

## Test plan
- [ ] Renovate creates its onboarding PR or starts creating dependency update PRs
EOF
)"
```

---

## PR 6: Release Automation

**Branch:** `tooling/release-automation`
**Base:** `main`

Adds a lightweight release workflow for the library. Since admin-apis has no binary artifacts, goreleaser is not applicable. Instead, this uses a tag-push-triggered workflow that creates a GitHub Release with auto-generated release notes from conventional commits.

### Task 6.1: Add release workflow

**Files:**
- Create: `.github/workflows/release.yaml`

- [ ] **Step 1: Create `.github/workflows/release.yaml`**

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: setup Go
        uses: actions/setup-go@v6
        with:
          go-version-file: go.mod

      - name: Verify build
        run: go build ./...

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
```

How it works:
- Push a tag (`git tag v1.2.3 && git push origin v1.2.3`) to trigger a release
- Verifies the code builds before creating the release
- Uses GitHub's auto-generated release notes (groups PRs by label since last tag)
- The existing `notify_repositories` workflow handles downstream signaling separately (triggered on push to `main`, not on release)

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yaml
git commit -m "feat: add release automation workflow

Creates GitHub Releases with auto-generated release notes when
a version tag is pushed. Library-appropriate alternative to goreleaser."
```

- [ ] **Step 3: Push and open PR**

```bash
git push -u origin tooling/release-automation
gh pr create --title "feat: add release automation" --body "$(cat <<'EOF'
## Summary
- Adds release workflow triggered by version tags (`v*`)
- Verifies build, then creates GitHub Release with auto-generated notes
- Library-appropriate: no goreleaser (no binary artifacts)
- Existing `notify_repositories` workflow handles downstream signaling independently

## Test plan
- [ ] After merging, push a tag: `git tag v0.1.0 && git push origin v0.1.0`
- [ ] Verify GitHub Release is created with correct notes
EOF
)"
```

---

## Execution Order & Dependencies

```
PR 1: golangci-lint config ──┐
PR 2: pre-commit hooks ──────┤ (uses golangci-lint, benefits from config existing)
PR 3: Taskfile ──────────────┤ (independent)
PR 4: CI modernization ──────┤ (independent, but pairs well with PR 1-2)
PR 5: Renovate ──────────────┤ (best after pre-commit config exists)
PR 6: Release automation ────┘ (independent)
```

**Recommended merge order:** PR 1 → PR 2 → PRs 3-6 (any order)

PRs 1 and 2 have a soft dependency (pre-commit hooks use golangci-lint, so `.golangci.yaml` should exist first). All others are independent.

**Stacked branching alternative:** If you prefer stacked branches, chain them: `main` → `tooling/golangci-lint` → `tooling/pre-commit` → `tooling/taskfile` → `tooling/ci-modernization` → `tooling/renovate` → `tooling/release-automation`. Merge bottom-up.
