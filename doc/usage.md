# gerrit-cli Usage

This document is generated from the actual source code and cross-checked against docs in `doc/`.

## Overview

`gerrit-cli` is a Gerrit code review command-line tool.

Top-level commands:

```text
gerrit
├── init
├── change (alias: changes)
│   ├── list (alias: ls)
│   ├── view
│   ├── comments
│   ├── review
│   ├── draft
│   ├── drafts
│   ├── submit
│   ├── abandon
│   ├── rebase
│   ├── diff
│   ├── patch
│   └── reviewers
│       ├── add
│       └── remove
└── version
```

---

## Global Flags

Available on all commands:

- `--config string` : config file path (default behavior uses `~/.gerrit/.config.yml`)
- `--profile string` : profile name (overrides `GERRIT_PROFILE`)
- `--debug` : enable debug output (HTTP request/response dump)

---

## Configuration

Initialize config interactively:

```bash
gerrit init
```

Config file: `~/.gerrit/.config.yml`

Multi-profile format:

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

Use a profile:

```bash
gerrit --profile oss change list
GERRIT_PROFILE=oss gerrit change list
```

---

## `change list` / `change ls`

List changes.

```bash
gerrit change list [flags]
```

Flags:

- `-s, --status string` (default: `open`)
- `--reviewer` (filter by `reviewer:self`)
- `-p, --project string`
- `-b, --branch string`
- `-q, --query string` (raw Gerrit query)
- `-n, --limit int` (default: `25`)
- `--plain`
- `--no-headers`
- `--json`

Examples:

```bash
gerrit change list
gerrit change list -s merged -p my/project -n 50
gerrit change ls --reviewer --json
```

Notes:
- If `--query` is set, structured filters are skipped.
- If `--project` is not provided, it falls back to configured default project.

---

## `change view <change-id>`

View change details.

```bash
gerrit change view <change-id> [flags]
```

Flags:

- `--files`
- `--patchsets`
- `--json`

Examples:

```bash
gerrit change view 3525
gerrit change view 3525 --patchsets
gerrit change view 3525 --files
```

---

## `change comments <change-id>`

View comments.

```bash
gerrit change comments <change-id> [flags]
```

Flags:

- `-p, --patchset int`
- `--all`
- `--json`

Examples:

```bash
gerrit change comments 3525
gerrit change comments 3525 -p 3 --all
```

Notes:
- Default output is unresolved comments only.

---

## `change review <change-id>`

Post a review (message, votes, inline comments).

```bash
gerrit change review <change-id> [flags]
```

Flags:

- `-p, --patchset int` (default: current)
- `-m, --message string`
- `--code-review int`
- `--verified int`
- `--label stringArray` (format: `Name=score`)
- `-c, --comment stringArray` (format: `file:line:message`)

Examples:

```bash
gerrit change review 3525 -m "LGTM" --code-review +1
gerrit change review 3525 --label "Quality=+1"
gerrit change review 3525 -c "src/main.go:42:Please rename this variable"
gerrit change review 3525 -m "See inline comments" -c "src/main.go:42:needs nil check"
```

Notes:
- `--comment` format must be exactly `file:line:message`.

---

## `change draft <change-id>`

Create draft inline comments.

```bash
gerrit change draft <change-id> [flags]
```

Flags:

- `-p, --patchset int` (default: current)
- `-c, --comment stringArray` (**required**, repeatable, format: `file:line:message`)

Examples:

```bash
gerrit change draft 3525 -c "src/main.go:42:test comment"
gerrit change draft 3525 -p 2 \
  -c "a.go:10:needs tests" \
  -c "b.go:20:typo"
```

---

## `change drafts <change-id>`

List draft comments.

```bash
gerrit change drafts <change-id> [flags]
```

Flags:

- `-p, --patchset int` (default: current)

Examples:

```bash
gerrit change drafts 3525
gerrit change drafts 3525 -p 2
```

---

## `change submit <change-id>`

Submit a change.

```bash
gerrit change submit <change-id>
```

Example:

```bash
gerrit change submit 3525
```

---

## `change abandon <change-id>`

Abandon a change.

```bash
gerrit change abandon <change-id> [flags]
```

Flags:

- `-m, --message string`

Example:

```bash
gerrit change abandon 3525 -m "obsolete"
```

---

## `change rebase <change-id>`

Rebase a change.

```bash
gerrit change rebase <change-id> [flags]
```

Flags:

- `-b, --base string`
- `--allow-conflicts`

Example:

```bash
gerrit change rebase 3525 --base main --allow-conflicts
```

---

## `change diff <change-id>`

View diffs.

```bash
gerrit change diff <change-id> [flags]
```

Flags:

- `-p, --patchset int` (default: current)
- `--base int`
- `-f, --file string`
- `-C, --context int` (default: `3`)
- `--no-color`
- `--stat`

Examples:

```bash
gerrit change diff 3525
gerrit change diff 3525 --stat
gerrit change diff 3525 -p 3 --base 1
gerrit change diff 3525 -f src/main.go -C 5
```

---

## `change patch <change-id>`

Download patch content.

```bash
gerrit change patch <change-id> [flags]
```

Flags:

- `-p, --patchset int` (default: current)
- `-o, --output string`

Examples:

```bash
gerrit change patch 3525
gerrit change patch 3525 -o change.patch
gerrit change patch 3525 -p 2 -o change-ps2.patch
```

---

## `change reviewers`

Manage reviewers.

### `change reviewers add <change-id> <account>`

Flags:

- `--cc` (add as CC instead of reviewer)

Examples:

```bash
gerrit change reviewers add 3525 user@example.com
gerrit change reviewers add 3525 user@example.com --cc
```

### `change reviewers remove <change-id> <account>`

Example:

```bash
gerrit change reviewers remove 3525 user@example.com
```

---

## `version`

```bash
gerrit version
```

---

## Source-Verified Notes / Caveats

1. `change patch --zip` is **not implemented** in current code.
2. Patchset selection generally passes numeric revision directly when `-p` is set; otherwise uses `current`.
3. `change comments` defaults to unresolved-only unless `--all` is specified.
4. Some write APIs may fail with `403` depending on Gerrit/nginx auth configuration (`--debug` helps diagnose).

---

## Troubleshooting

Show request/response details:

```bash
gerrit change view 3525 --debug
```

Reinitialize config/profile:

```bash
gerrit init
```
