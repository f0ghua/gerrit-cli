# gerrit-cli

A lightweight command-line tool for Gerrit code review workflows.

`gerrit-cli` helps you query changes, inspect details, read/post comments, manage reviewers, submit/abandon/rebase changes, and work with diffs/patches directly from your terminal.

## Features

- Interactive setup with profile-based configuration (`gerrit init`)
- Change listing with filters (status, project, branch, reviewer, raw query)
- Rich change inspection (labels, reviewers, files, patchsets)
- Comment workflows:
  - View comments (unresolved by default, optional all)
  - Post reviews with votes/labels/inline comments
  - Create/list draft inline comments
- Diff and patch workflows:
  - File-level or full change diffs
  - Diffstat output
  - Patch download to stdout or file
- Change lifecycle operations:
  - Submit
  - Abandon
  - Rebase
  - Reviewer add/remove (including CC mode)
- JSON output options for automation-friendly workflows

## Installation

### Prerequisites

- Go `1.23.0` or newer

### Build from source

```bash
make build
```

Binary output:

```text
bin/gerrit
```

### Install into your Go bin path

```bash
make install
```

or

```bash
go install ./cmd/gerrit
```

## Quick Start

1. Initialize configuration:

```bash
gerrit init
```

2. List your open changes:

```bash
gerrit change list
```

3. View a change:

```bash
gerrit change view <change-id>
```

## Configuration

Default config file:

```text
~/.gerrit/.config.yml
```

### Profile format

```yaml
default: work
profiles:
  work:
    server: http://gerrit.example.com
    username: alice
    password: "token"
    project: my/project
    no_auth_prefix: false
  oss:
    server: https://gerrit-review.googlesource.com
    username: alice
    password: "token"
```

### Profile selection precedence

1. `--profile`
2. `GERRIT_PROFILE`
3. `default` in config

Examples:

```bash
gerrit --profile oss change list
GERRIT_PROFILE=oss gerrit change list
```

### Environment variables

- `GERRIT_PROFILE`: active profile when `--profile` is not provided
- `GERRIT_HTTP_PASSWORD`: password fallback when config password is empty
- Proxy settings are read from standard environment variables via Go HTTP transport

## Command Overview

### Top-level commands

- `gerrit init`
- `gerrit change` (alias: `changes`)
- `gerrit version`

### Change subcommands

- `list` (`ls`)
- `view <change-id>`
- `comments <change-id>`
- `review <change-id>`
- `draft <change-id>`
- `drafts <change-id>`
- `submit <change-id>`
- `abandon <change-id>`
- `rebase <change-id>`
- `diff <change-id>`
- `patch <change-id>`
- `reviewers add <change-id> <account>`
- `reviewers remove <change-id> <account>`

### Global flags

Available on all commands:

- `--config string` (config file path)
- `--profile string` (profile name)
- `--debug` (HTTP request/response dump)

## Usage Examples

```bash
# list open changes (default status=open)
gerrit change list

# filtered list
gerrit change list -s merged -p my/project -n 50
gerrit change ls --reviewer --json

# inspect a change
gerrit change view 3525 --patchsets
gerrit change view 3525 --files

# comments
gerrit change comments 3525
gerrit change comments 3525 -p 3 --all

# post review
gerrit change review 3525 -m "LGTM" --code-review +1
gerrit change review 3525 --label "Quality=+1"
gerrit change review 3525 -c "src/main.go:42:Please rename this variable"

# draft workflow
gerrit change draft 3525 -c "src/main.go:42:needs nil check"
gerrit change drafts 3525

# diff and patch
gerrit change diff 3525 --stat
gerrit change diff 3525 -f src/main.go -C 5
gerrit change patch 3525 -o change.patch

# lifecycle + reviewers
gerrit change submit 3525
gerrit change abandon 3525 -m "obsolete"
gerrit change rebase 3525 --base main --allow-conflicts
gerrit change reviewers add 3525 user@example.com --cc
gerrit change reviewers remove 3525 user@example.com
```

## Project Structure

```text
cmd/gerrit/                # CLI entrypoint
internal/cmd/              # cobra command tree and subcommands
internal/config/           # config loading/saving and profile model
internal/view/             # terminal rendering for changes/comments/etc.
internal/cmdutil/          # shared command helpers
api/                       # client factory and config-to-client wiring
pkg/gerrit/                # Gerrit REST client + API operations + types
pkg/query/                 # Gerrit query builder
```

## Development

### Common tasks

```bash
make build
make install
make clean
```

### Dependencies

Main direct dependencies:

- `github.com/spf13/cobra`
- `github.com/spf13/viper`
- `github.com/fatih/color`
- `gopkg.in/yaml.v3`

### Tests

There are currently no Go test files (`*_test.go`) in the repository.

## Notes and Caveats

- Most commands require initialized configuration; run `gerrit init` first.
- `change comments` shows unresolved comments by default; use `--all` to include resolved comments.
- `change patch --zip` is not currently implemented.
- `--debug` dumps raw HTTP request/response data and may expose sensitive headers/tokens in terminal output; use carefully.

## License

No license file is currently present in this repository.