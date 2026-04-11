# admin-apis

Go library providing shared API types and license definitions for the [DevPod](https://github.com/skevetter/devpod) ecosystem.

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [Task](https://taskfile.dev/) (task runner)

## Usage

```bash
go get github.com/skevetter/admin-apis
```

## Development

```bash
# List available tasks
task --list

# Run linter
task lint

# Generate API types and OpenAPI definitions
task gen

# Verify generated files are up to date
task check:gen

# Check struct memory alignment
task check:structalign
```

## Features

Feature definitions live in `definitions/features.yaml`. To add, remove, or edit features:

1. Edit `definitions/features.yaml`
2. Run `task gen`
3. Commit the generated changes

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE).
