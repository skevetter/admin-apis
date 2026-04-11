# admin-apis Tooling Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring admin-apis tooling in line with devpod ecosystem standards — pre-commit, linting, Taskfile, dependency management (renovate), and automated releases (release-please).

**Architecture:** Eight sequential PRs, each independently mergeable in order. Each PR adds one concern. PR 0 migrates the module path first since all other work depends on it. Config files are copied from provider-aws/devpod as the canonical source and adapted for this library repo.

**Tech Stack:** Go 1.24, pre-commit, golangci-lint v2, Taskfile v3, Renovate, release-please, GitHub Actions

---

## File Map

**Create:**
- `.pre-commit-config.yaml` — pre-commit hook definitions
- `commitlint.config.mjs` — conventional commit rules
- `requirements.txt` — Python deps for CI pre-commit
- `.golangci.yaml` — linter/formatter configuration
- `Taskfile.yml` — task runner (replaces Justfile)
- `renovate.json` — dependency automation config
- `release-please-config.json` — release-please settings
- `.release-please-manifest.json` — version tracking
- `.github/workflows/commit.yml` — commit checks on PRs
- `.github/workflows/pre-commit.yml` — pre-commit CI
- `.github/workflows/release.yml` — release-please automation
- `.github/workflows/workflow-approval.yml` — auto-approve workflows

**Modify:**
- `go.mod` — change module path from `github.com/loft-sh/admin-apis` to `github.com/skevetter/admin-apis`
- `pkg/**/*.go` — update all import paths from `loft-sh` to `skevetter`
- `hack/**/*.go` — update import paths
- `zz_generated.*` files — regenerated with new module path
- `.github/workflows/lint.yaml` — update action versions, add concurrency
- `.editorconfig` — remove Justfile section
- `pkg/**/*.go` — lint/format fixes (automated)

**Delete:**
- `Justfile` — replaced by Taskfile
- `.github/workflows/go.yml` — go generate check moves to Taskfile
- `.github/workflows/notify_repositories.yaml` — replaced by renovate

---

## Task 0: Module Path Migration (PR 0)

**Files:**
- Modify: `go.mod` (module path)
- Modify: `hack/gen-features/main.go` (import paths)
- Modify: `pkg/licenseapi/doc.go` (openapi model package annotation)
- Modify: `pkg/util/features/features.go` (import paths)
- Regenerate: `pkg/licenseapi/zz_generated.model_name.go` (model name strings)
- Regenerate: `pkg/licenseapi/zz_generated.deepcopy.go`
- Regenerate: `pkg/licenseapi/zz_generated.openapi.go`
- Modify: `Justfile` (gen command references loft-sh module path)

- [ ] **Step 1: Create branch**

```bash
cd ~/ws/devpod/admin-apis
git checkout main
git checkout -b chore/migrate-module-path
```

- [ ] **Step 2: Update go.mod module path**

Change `module github.com/loft-sh/admin-apis` to `module github.com/skevetter/admin-apis` in `go.mod`.

- [ ] **Step 3: Update all Go import paths**

Replace `github.com/loft-sh/admin-apis` with `github.com/skevetter/admin-apis` in all `.go` files:

```bash
find . -name '*.go' -exec sed -i 's|github.com/loft-sh/admin-apis|github.com/skevetter/admin-apis|g' {} +
```

- [ ] **Step 4: Update Justfile gen command**

The openapi-gen command in `Justfile` references `github.com/loft-sh/admin-apis/pkg/licenseapi`. Update both `--output-pkg` and the input package argument to use `github.com/skevetter/admin-apis/pkg/licenseapi`.

- [ ] **Step 5: Run go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 6: Regenerate generated files**

```bash
just gen
```

This regenerates `zz_generated.deepcopy.go`, `zz_generated.openapi.go`, and `zz_generated.model_name.go` with the new module path.

- [ ] **Step 7: Verify build**

```bash
go build ./...
go vet ./...
```

Expected: both pass with zero errors.

- [ ] **Step 8: Verify no loft-sh references remain**

```bash
grep -r "loft-sh" --include="*.go" --include="go.mod" --include="Justfile" .
```

Expected: zero matches.

- [ ] **Step 9: Commit**

```bash
git add -u
git commit -m "chore: migrate module path to skevetter/admin-apis

Change Go module from github.com/loft-sh/admin-apis to
github.com/skevetter/admin-apis. Update all imports, generated
files, and code generation commands."
```

- [ ] **Step 10: Push and create PR**

```bash
git push -u origin chore/migrate-module-path
gh pr create --title "chore: migrate module path to skevetter/admin-apis" --body "$(cat <<'EOF'
## Summary
- Change Go module path from `github.com/loft-sh/admin-apis` to `github.com/skevetter/admin-apis`
- Update all import paths in source and generated files
- Update code generation commands in Justfile

## Test plan
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
- [ ] No remaining `loft-sh` references in Go files, go.mod, or Justfile
EOF
)"
```

---

## Task 1: Pre-commit Setup (PR 1)

**Files:**
- Create: `.pre-commit-config.yaml`
- Create: `commitlint.config.mjs`
- Create: `requirements.txt`

- [ ] **Step 1: Create `.pre-commit-config.yaml`**

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

- [ ] **Step 2: Create `commitlint.config.mjs`**

```javascript
export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "type-enum": [
      2,
      "always",
      [
        "build",
        "ci",
        "docs",
        "feat",
        "fix",
        "perf",
        "refactor",
        "style",
        "test",
        "chore",
        "revert",
        "bump",
        "fixup",
      ],
    ],
    "body-max-line-length": [0, "always", Infinity],
  },
}
```

- [ ] **Step 3: Create `requirements.txt`**

```
pre_commit>=4.3.0
```

- [ ] **Step 4: Install pre-commit hooks locally**

Run:
```bash
cd ~/ws/devpod/admin-apis
pip install pre-commit
pre-commit install
pre-commit install --hook-type commit-msg
```

- [ ] **Step 5: Run pre-commit on all files and fix violations**

Run:
```bash
pre-commit run --all-files
```

Review output. Most hooks auto-fix (trailing whitespace, EOF, shfmt). For any that don't auto-fix, manually resolve. Re-run until clean:

```bash
pre-commit run --all-files
```

Expected: all hooks pass.

- [ ] **Step 6: Create branch and commit**

```bash
git checkout -b chore/add-pre-commit
git add .pre-commit-config.yaml commitlint.config.mjs requirements.txt
git add -u  # any auto-fixed files
git commit -m "chore: add pre-commit hooks

Add pre-commit configuration matching devpod ecosystem standards.
Includes commitlint, shellcheck, shfmt, actionlint, gitleaks,
commitizen, and golangci-lint hooks."
```

- [ ] **Step 7: Push and create PR**

```bash
git push -u origin chore/add-pre-commit
gh pr create --title "chore: add pre-commit hooks" --body "$(cat <<'EOF'
## Summary
- Add `.pre-commit-config.yaml` with standard devpod hooks
- Add `commitlint.config.mjs` for conventional commit enforcement
- Add `requirements.txt` for CI pre-commit dependency
- Fix any existing violations caught by pre-commit

## Test plan
- [ ] `pre-commit run --all-files` passes locally
- [ ] CI checks pass
EOF
)"
```

---

## Task 2: Golangci-lint Config (PR 2)

**Depends on:** PR 1 merged (pre-commit includes golangci-lint hook that references this config)

**Files:**
- Create: `.golangci.yaml`
- Modify: `pkg/**/*.go` (automated lint/format fixes)

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

- [ ] **Step 2: Run golangci-lint with auto-fix**

Run:
```bash
golangci-lint run --fix ./...
```

This auto-fixes formatting (gci, gofumpt, goimports, golines) and some linter issues.

- [ ] **Step 3: Run golangci-lint again to check remaining issues**

Run:
```bash
golangci-lint run ./...
```

Manually fix any remaining violations. Common ones to expect:
- `godot` — missing periods at end of comments
- `lll` — long lines
- `funlen` — long functions
- `cyclop` — complex functions
- `errcheck` — unchecked errors

Re-run until clean.

- [ ] **Step 4: Verify pre-commit still passes**

Run:
```bash
pre-commit run --all-files
```

Expected: all hooks pass.

- [ ] **Step 5: Create branch and commit**

```bash
git checkout main
git pull
git checkout -b chore/add-golangci-lint-config
git add .golangci.yaml
git add -u  # lint-fixed Go files
git commit -m "chore: add golangci-lint configuration

Add .golangci.yaml with standard devpod linter/formatter config.
Fix all existing lint violations across the codebase."
```

- [ ] **Step 6: Push and create PR**

```bash
git push -u origin chore/add-golangci-lint-config
gh pr create --title "chore: add golangci-lint configuration" --body "$(cat <<'EOF'
## Summary
- Add `.golangci.yaml` with 23 linters and 4 formatters matching devpod standards
- Fix all existing lint/format violations

## Test plan
- [ ] `golangci-lint run ./...` passes with zero issues
- [ ] `pre-commit run --all-files` passes
- [ ] `go build ./...` succeeds
EOF
)"
```

---

## Task 3: Taskfile (PR 3)

**Files:**
- Create: `Taskfile.yml`
- Delete: `Justfile`
- Modify: `.editorconfig` (remove Justfile section)

- [ ] **Step 1: Create `Taskfile.yml`**

```yaml
version: "3"

tasks:
  lint:
    desc: Run golangci-lint for all packages
    cmds:
      - golangci-lint run {{.CLI_ARGS}}

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

  gen:check:
    desc: Check that generated files are up to date
    cmds:
      - task: gen
      - |
        if [ -n "$(git diff)" ]; then
          echo "Generated files are out of date. Run 'task gen' and commit the changes."
          exit 1
        fi

  check-structalign:
    desc: Check struct memory alignment and print potential improvements
    cmds:
      - go run github.com/dkorunic/betteralign/cmd/betteralign@latest {{.CLI_ARGS}} ./...
```

- [ ] **Step 2: Verify Taskfile works**

Run:
```bash
task lint
task --list
```

Expected: `task lint` runs golangci-lint successfully. `task --list` shows all four tasks.

- [ ] **Step 3: Delete Justfile**

```bash
rm Justfile
```

- [ ] **Step 4: Update `.editorconfig` — remove Justfile section**

Remove these lines from `.editorconfig`:

```
[Justfile]
indent_style = space
indent_size = 2
```

- [ ] **Step 5: Create branch and commit**

```bash
git checkout main
git pull
git checkout -b chore/replace-justfile-with-taskfile
git add Taskfile.yml .editorconfig
git rm Justfile
git commit -m "chore: replace Justfile with Taskfile

Migrate task definitions to Taskfile.yml for consistency with other
devpod Go repos. Adds gen:check task for CI go-generate freshness check."
```

- [ ] **Step 6: Push and create PR**

```bash
git push -u origin chore/replace-justfile-with-taskfile
gh pr create --title "chore: replace Justfile with Taskfile" --body "$(cat <<'EOF'
## Summary
- Add `Taskfile.yml` with lint, gen, gen:check, and check-structalign tasks
- Remove `Justfile`
- Update `.editorconfig` to remove Justfile section

## Test plan
- [ ] `task lint` runs successfully
- [ ] `task gen:check` passes (generated files are up to date)
- [ ] `task --list` shows all tasks
EOF
)"
```

---

## Task 4: Dependency Updates (PR 4)

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Update Go module dependencies**

Run:
```bash
cd ~/ws/devpod/admin-apis
go get -u ./...
go mod tidy
```

- [ ] **Step 2: Verify build still works**

Run:
```bash
go build ./...
```

Expected: builds successfully with no errors.

- [ ] **Step 3: Check for breaking changes**

Run:
```bash
go vet ./...
```

If there are compilation or vet errors from major version bumps, fix them.

- [ ] **Step 4: Run linter to catch any issues from dependency changes**

Run:
```bash
golangci-lint run ./...
```

Fix any new issues introduced by dependency updates.

- [ ] **Step 5: Create branch and commit**

```bash
git checkout main
git pull
git checkout -b chore/update-dependencies
git add go.mod go.sum
git add -u  # any Go files fixed due to API changes
git commit -m "chore: update Go module dependencies

Update all direct and indirect dependencies to latest versions."
```

- [ ] **Step 6: Push and create PR**

```bash
git push -u origin chore/update-dependencies
gh pr create --title "chore: update Go module dependencies" --body "$(cat <<'EOF'
## Summary
- Update all Go module dependencies to latest versions
- Run go mod tidy

## Test plan
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
- [ ] `golangci-lint run ./...` passes
EOF
)"
```

---

## Task 5: Renovate (PR 5)

**Files:**
- Create: `renovate.json`
- Delete: `.github/workflows/notify_repositories.yaml`

- [ ] **Step 1: Create `renovate.json`**

```json
{
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "automerge": false,
    "automergeType": "pr",
    "automergeStrategy": "squash",
    "extends": [
        "config:recommended"
    ],
    "pre-commit": {
        "enabled": true
    },
    "dependencyDashboard": false,
    "prConcurrentLimit": 1,
    "platformAutomerge": true,
    "postUpdateOptions": [
        "gomodTidy"
    ],
    "packageRules": [
        {
            "matchUpdateTypes": [
                "major"
            ],
            "automerge": false,
            "prPriority": -1
        },
        {
            "matchUpdateTypes": [
                "minor",
                "patch"
            ],
            "prPriority": 10
        },
        {
            "matchManagers": [
                "github-actions"
            ],
            "automerge": true,
            "prPriority": 5
        },
        {
            "matchManagers": [
                "pre-commit"
            ],
            "automerge": true,
            "prPriority": 4
        },
        {
            "matchManagers": [
                "gomod"
            ],
            "automerge": true,
            "prPriority": 3
        }
    ]
}
```

- [ ] **Step 2: Delete `notify_repositories.yaml`**

```bash
rm .github/workflows/notify_repositories.yaml
```

- [ ] **Step 3: Create branch and commit**

```bash
git checkout main
git pull
git checkout -b chore/add-renovate
git add renovate.json
git rm .github/workflows/notify_repositories.yaml
git commit -m "chore: add Renovate for dependency management

Add renovate.json with standard devpod config. Auto-merges minor/patch
for gomod, GitHub Actions, and pre-commit. Manual review for major bumps.

Remove notify_repositories.yaml — downstream repos will be notified of
new versions via Renovate dependency PRs instead."
```

- [ ] **Step 4: Push and create PR**

```bash
git push -u origin chore/add-renovate
gh pr create --title "chore: add Renovate for dependency management" --body "$(cat <<'EOF'
## Summary
- Add `renovate.json` with standard devpod config
- Remove `notify_repositories.yaml` (replaced by Renovate for downstream notifications)

## Test plan
- [ ] Verify Renovate GitHub App is installed on this repo
- [ ] After merge, Renovate should create its onboarding PR or start creating dependency PRs
EOF
)"
```

---

## Task 6: Release Automation (PR 6)

**Files:**
- Create: `.github/workflows/release.yml`
- Create: `release-please-config.json`
- Create: `.release-please-manifest.json`

- [ ] **Step 1: Create `release-please-config.json`**

```json
{
    "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
    "release-type": "go",
    "packages": {
        ".": {
            "changelog-path": "CHANGELOG.md",
            "bump-minor-pre-major": true,
            "bump-patch-for-minor-pre-major": true
        }
    }
}
```

- [ ] **Step 2: Create `.release-please-manifest.json`**

```json
{
    ".": "0.0.0"
}
```

This tells release-please the current version. Starting at 0.0.0 means the first release PR will bump to the appropriate version based on conventional commits (feat → 0.1.0 or 1.0.0 depending on config).

- [ ] **Step 3: Create `.github/workflows/release.yml`**

```yaml
name: Release

on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-latest
    steps:
      - uses: googleapis/release-please-action@v4
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
```

- [ ] **Step 4: Create branch and commit**

```bash
git checkout main
git pull
git checkout -b chore/add-release-please
git add release-please-config.json .release-please-manifest.json .github/workflows/release.yml
git commit -m "ci: add release-please for automated releases

Configure release-please to create release PRs from conventional commits.
Merging a release PR creates a GitHub Release with semver tag and changelog."
```

- [ ] **Step 5: Push and create PR**

```bash
git push -u origin chore/add-release-please
gh pr create --title "ci: add release-please for automated releases" --body "$(cat <<'EOF'
## Summary
- Add `release-please-config.json` and `.release-please-manifest.json`
- Add `.github/workflows/release.yml` triggered on push to main
- On merge to main, release-please creates/updates a release PR with changelog
- Merging the release PR tags and creates a GitHub Release

## Test plan
- [ ] After merge, push a `feat:` commit to main and verify release-please creates a release PR
- [ ] Merge the release PR and verify a GitHub Release is created with correct tag
EOF
)"
```

---

## Task 7: CI Modernization (PR 7)

**Depends on:** PRs 1-3 merged (references pre-commit, golangci-lint config, Taskfile)

**Files:**
- Create: `.github/workflows/commit.yml`
- Create: `.github/workflows/pre-commit.yml`
- Create: `.github/workflows/workflow-approval.yml`
- Modify: `.github/workflows/lint.yaml`
- Delete: `.github/workflows/go.yml`

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

- [ ] **Step 2: Create `.github/workflows/pre-commit.yml`**

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

- [ ] **Step 3: Create `.github/workflows/workflow-approval.yml`**

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

- [ ] **Step 4: Update `.github/workflows/lint.yaml`**

Replace the entire file with:

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
      - uses: jlumbroso/free-disk-space@v1.3.1

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

Note: lint.yaml already has v6/v9 versions from an earlier update. This step ensures it also has the `free-disk-space` step and consistent formatting.

- [ ] **Step 5: Delete `.github/workflows/go.yml`**

```bash
rm .github/workflows/go.yml
```

The `go generate` freshness check is now available as `task gen:check` and can be run in the pre-commit or lint workflow if needed.

- [ ] **Step 6: Create branch and commit**

```bash
git checkout main
git pull
git checkout -b chore/modernize-ci
git add .github/workflows/commit.yml .github/workflows/pre-commit.yml .github/workflows/workflow-approval.yml .github/workflows/lint.yaml
git rm .github/workflows/go.yml
git commit -m "ci: modernize GitHub Actions workflows

Add commit verification, pre-commit CI, and workflow auto-approval.
Update lint workflow with free-disk-space step.
Remove go.yml — go generate check is now available via 'task gen:check'."
```

- [ ] **Step 7: Push and create PR**

```bash
git push -u origin chore/modernize-ci
gh pr create --title "ci: modernize GitHub Actions workflows" --body "$(cat <<'EOF'
## Summary
- Add `commit.yml` — signed commit + commitlint checks on PRs
- Add `pre-commit.yml` — run all pre-commit hooks in CI
- Add `workflow-approval.yml` — auto-approve workflow runs
- Update `lint.yaml` with free-disk-space step
- Remove `go.yml` (go generate check available via `task gen:check`)

## Test plan
- [ ] All new workflows trigger correctly on PR
- [ ] `pre-commit.yml` runs all hooks successfully
- [ ] `commit.yml` validates conventional commit messages
- [ ] `lint.yaml` runs golangci-lint successfully
EOF
)"
```
