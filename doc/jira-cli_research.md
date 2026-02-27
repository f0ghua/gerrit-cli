# jira-cli Codebase Research Report

## 1. Project Overview

**Repository:** `github.com/ankitpokhrel/jira-cli`
**Language:** Go (CGO disabled by default)
**CLI Framework:** Cobra + Viper
**Purpose:** An interactive command-line interface for Jira, supporting both Cloud and on-premise (Local) installations with multi-version API support (v1, v2, v3).

---

## 2. Directory Structure

```
jira-cli/
├── cmd/jira/main.go              # Entry point (minimal bootstrap)
├── api/client.go                 # API proxy layer (v2/v3 routing)
├── pkg/                          # Public, reusable packages
│   ├── jira/                     # Core Jira HTTP client & API operations
│   ├── adf/                      # Atlassian Document Format parser
│   ├── md/                       # Markdown <-> Jira wiki conversion
│   ├── jql/                      # JQL query builder (fluent API)
│   ├── tui/                      # Terminal UI components (tview-based)
│   ├── browser/                  # Open URLs in system browser
│   ├── netrc/                    # .netrc file reader
│   └── surveyext/                # Extensions for survey prompts
├── internal/                     # Private packages
│   ├── cmd/                      # All CLI command implementations
│   │   ├── root/                 # Root command + config init
│   │   ├── issue/                # Issue CRUD + subcommands
│   │   ├── epic/                 # Epic management
│   │   ├── sprint/               # Sprint management
│   │   ├── board/                # Board listing
│   │   ├── project/              # Project listing
│   │   ├── release/              # Release/version listing
│   │   ├── init/                 # Interactive config wizard
│   │   ├── me/                   # Current user info
│   │   ├── open/                 # Open issue in browser
│   │   ├── serverinfo/           # Server info
│   │   ├── version/              # CLI version
│   │   ├── completion/           # Shell completion
│   │   └── man/                  # Man page generation
│   ├── cmdcommon/                # Shared create/edit logic & flags
│   ├── cmdutil/                  # Utilities (spinners, error handling, datetime)
│   ├── config/                   # Config file generation wizard
│   ├── query/                    # Flag-to-JQL query translation
│   └── view/                     # Output rendering (table, markdown, plain)
├── Makefile
├── go.mod / go.sum
└── Dockerfile / docker-compose.yml
```

---

## 3. Architecture & Design Patterns

### 3.1 Layered Architecture

The codebase follows a clean 4-layer separation:

```
┌─────────────────────────────────────────────┐
│  CLI Layer (internal/cmd/*)                  │  Cobra commands, flag parsing, user interaction
├─────────────────────────────────────────────┤
│  Proxy Layer (api/client.go)                │  v2/v3 routing based on installation type
├─────────────────────────────────────────────┤
│  Client Layer (pkg/jira/*)                  │  HTTP client, request/response handling
├─────────────────────────────────────────────┤
│  View Layer (internal/view/*)               │  Output formatting (TUI, plain, JSON, CSV)
└─────────────────────────────────────────────┘
```

### 3.2 Key Design Patterns

1. **Proxy Pattern** (`api/client.go`): Central routing layer that dispatches to v2 or v3 API based on `installation` config value. Every API call goes through a `Proxy*` function:
   - `ProxyCreate()`, `ProxyGetIssue()`, `ProxySearch()`, `ProxyAssignIssue()`
   - `ProxyUserSearch()`, `ProxyTransitions()`, `ProxyWatchIssue()`

2. **Singleton** (`api/client.go`): The Jira client is lazily initialized once and cached in a package-level `jiraClient` variable.

3. **Functional Options** (`pkg/jira/client.go`): Client configuration via `ClientFunc` closures:
   ```go
   jira.NewClient(config, jira.WithTimeout(15*time.Second), jira.WithInsecureTLS(false))
   ```

4. **Builder Pattern** (`pkg/jql/jql.go`): Fluent JQL query construction:
   ```go
   jql.NewJQL("PROJ").FilterBy("type", "Bug").In("status", "Open", "In Progress").OrderBy("created", "DESC")
   ```

5. **Custom JSON Marshaling** (`pkg/jira/create.go`, `pkg/jira/edit.go`): `createFieldsMarshaler` and `editFieldsMarshaler` implement custom `MarshalJSON()` to handle dynamic fields (epic field names, custom fields) that can't be expressed with static struct tags.

6. **Flag-to-Query Translation** (`internal/query/`): `FlagParser` interface abstracts cobra flag access, enabling testable query construction from CLI flags.

---

## 4. API Client Design (`pkg/jira/`)

### 4.1 HTTP Client (`client.go`)

The `Client` struct wraps Go's `http.Client` with:
- Three API base URLs:
  - v3: `/rest/api/3` (Cloud default)
  - v2: `/rest/api/2` (Local/on-premise)
  - v1: `/rest/agile/1.0` (Agile endpoints: boards, sprints)
- Methods for each HTTP verb + API version:
  - `Get()`, `GetV2()`, `GetV1()`
  - `Post()`, `PostV2()`, `PostV1()`
  - `Put()`, `PutV2()`, `PutV1()`
  - `DeleteV2()`
- TLS 1.2 minimum, proxy from environment, configurable timeout (default 15s)
- Debug mode dumps full request/response via `httputil`

### 4.2 Authentication

Three auth types supported (`types.go`):

| Type | Header | Use Case |
|------|--------|----------|
| `basic` | `Authorization: Basic base64(login:token)` | Cloud (email + API token), Local (username + password) |
| `bearer` | `Authorization: Bearer <token>` | Personal Access Token (PAT) |
| `mtls` | Client certificate + optional Bearer | Mutual TLS with CA cert, client cert, client key |

Token resolution priority (highest to lowest):
1. `JIRA_API_TOKEN` environment variable
2. `.netrc` file entry matching server/login
3. System keyring (`jira-cli` service name)

### 4.3 API Operations

All operations in `pkg/jira/` follow a consistent pattern:
1. Build path string
2. Marshal request body (if POST/PUT)
3. Call versioned HTTP method
4. Check for nil response / unexpected status code
5. Decode JSON response into typed struct

**Issue operations** (`issue.go`):
- `GetIssue()` / `GetIssueV2()` - Fetch with ADF conversion for v3
- `GetIssueRaw()` / `GetIssueV2Raw()` - Raw JSON response
- `AssignIssue()` / `AssignIssueV2()` - Uses `accountId` (Cloud) vs `name` (Local)
- `LinkIssue()` / `UnlinkIssue()` / `GetLinkID()`
- `AddIssueComment()` - Supports internal comments via properties
- `AddIssueWorklog()` - With optional estimate adjustment
- `WatchIssue()` / `WatchIssueV2()`
- `RemoteLinkIssue()` - Add external links
- `GetField()` - Fetch all configured fields
- `GetIssueLinkTypes()` - Available link types

**Create** (`create.go`):
- `Create()` / `CreateV2()` - POST to `/issue`
- `CreateRequest` struct with all fields including custom fields
- Custom JSON marshaler handles dynamic epic field name and custom fields
- Markdown-to-Jira conversion for description body
- Supports `ForProjectType()` and `ForInstallationType()` to adjust behavior

**Edit** (`edit.go`):
- `Edit()` - PUT to `/issue/{key}` (always v2)
- Uses `update` semantics (add/remove for labels, components, versions)
- Prefix `-` on values means "remove" (e.g., `-label1` removes label1)
- Custom fields handled via custom marshaler

**Search** (`search.go`):
- `Search()` - v3 endpoint `/search/jql` with `fields=*all`
- `SearchV2()` - v2 endpoint `/search` with pagination (`startAt`, `maxResults`)

**Transitions** (`transition.go`):
- `Transitions()` / `TransitionsV2()` - GET available transitions
- `Transition()` - POST to move issue state, supports comment and field updates

**Sprint** (`sprint.go`):
- `Sprints()` - List sprints for a board (v1 agile API)
- `GetSprint()` / `EndSprint()` - Single sprint operations
- `SprintsInBoards()` - Concurrent fetch across multiple boards
- `SprintIssues()` - Issues in a sprint with JQL filter
- `SprintIssuesAdd()` - Add issues to sprint
- `lastNSprints()` - Workaround for Jira's ascending-only sort

**Board** (`board.go`): `Boards()`, `BoardSearch()` via v1 agile API

**Epic** (`epic.go`): `EpicIssues()`, `EpicIssuesAdd()`, `EpicIssuesRemove()` via v1 agile API

**Project** (`project.go`): `Project()` lists all projects via v2 API

**Other**: `Me()` (`me.go`), `ServerInfo()` (`serverinfo.go`), `GetCreateMeta()` (`createmeta.go`), `UserSearch()` (`user.go`), `Release()` (`release.go`)

### 4.4 Error Handling

- `ErrNoResult` - No results found
- `ErrEmptyResponse` - Nil response from server
- `ErrUnexpectedResponse` - Non-expected HTTP status code, carries parsed Jira error body
- `ErrMultipleFailed` - Grouped errors from batch operations

---

## 5. Command Structure

### 5.1 Root Command (`internal/cmd/root/root.go`)

Persistent flags: `--config`, `--project`, `--debug`

Config resolution:
1. `--config` flag
2. `JIRA_CONFIG_FILE` env var
3. `~/.jira/.config.yml` (default)

Viper binds env vars with `JIRA_` prefix and auto-reads config file.

`PersistentPreRun` checks for API token (except for `init`, `help`, `version`, `completion`, `man`).

### 5.2 Command Hierarchy

```
jira
├── init                          # Interactive config wizard
├── issue
│   ├── list (ls, search)         # List/search with JQL, plain/TUI/JSON/CSV output
│   ├── create                    # Create with interactive or non-interactive mode
│   ├── view                      # View issue details
│   ├── edit                      # Edit fields (add/remove semantics)
│   ├── delete                    # Delete issue
│   ├── assign                    # Assign to user
│   ├── move                      # Transition issue state
│   ├── link                      # Link two issues
│   ├── unlink                    # Remove link
│   ├── clone                     # Clone issue
│   ├── comment add               # Add comment (supports internal)
│   ├── watch                     # Watch/unwatch
│   └── worklog add               # Log time
├── epic
│   ├── list                      # List epics
│   ├── create                    # Create epic
│   ├── add                       # Add issue to epic
│   └── remove                    # Remove issue from epic
├── sprint
│   ├── list                      # List sprints (with issue sub-list)
│   ├── add                       # Add issue to sprint
│   └── close                     # Close sprint
├── board list                    # List boards
├── project list                  # List projects
├── release list                  # List releases/versions
├── open                          # Open issue in browser
├── me                            # Current user info
├── serverinfo                    # Server info
├── version                       # CLI version
├── completion                    # Shell completion (bash/zsh/fish)
└── man                           # Generate man pages
```

### 5.3 Issue List Command (representative example)

`internal/cmd/issue/list/list.go` demonstrates the typical command flow:

1. Parse flags into `query.IssueParams`
2. Build JQL via `query.NewIssue()` which uses `pkg/jql` builder
3. Call `api.ProxySearch()` (routes to v2 or v3)
4. Render via `view.IssueList` (supports TUI, plain table, CSV, raw JSON)

Flags include: `--type`, `--status`, `--priority`, `--assignee`, `--reporter`, `--label`, `--created`, `--updated`, `--jql`, `--paginate`, `--plain`, `--csv`, `--raw`, `--columns`, `--no-headers`, `--no-truncate`, `--delimiter`

---

## 6. Configuration System

### 6.1 Config File (`internal/config/generator.go`)

Location: `~/.jira/.config.yml`

Interactive wizard (`jira init`) flow:
1. Choose installation type (Cloud / Local)
2. Choose auth type for Local (basic / bearer / mtls)
3. Configure mTLS certs if applicable
4. Enter server URL and login credentials
5. Verify credentials via `Me()` API call
6. Fetch server version for Local installations
7. Select default project (from fetched list)
8. Select default board (with search capability)
9. Auto-configure issue types and custom fields from metadata

### 6.2 Config Structure

```yaml
installation: Cloud|Local
server: https://company.atlassian.net
login: user@example.com
auth_type: basic|bearer|mtls
project:
  key: PROJ
  type: software
board:
  id: 1
  name: "Board Name"
  type: scrum
epic:
  name: customfield_10011    # Dynamic field IDs
  link: customfield_10014
issue:
  types:
    - id: "10001"
      name: "Bug"
      handle: "Bug"
      subtask: false
  fields:
    custom:
      - name: "Story Points"
        key: "customfield_10016"
        schema:
          datatype: number
timezone: America/New_York
mtls:                         # Only for mTLS auth
  ca_cert: /path/to/ca.crt
  client_cert: /path/to/client.crt
  client_key: /path/to/client.key
```

---

## 7. Content Format Handling

### 7.1 Atlassian Document Format (`pkg/adf/`)

Jira v3 API uses ADF (a structured JSON document format) instead of plain text for descriptions and comments. The `adf` package:
- Parses ADF nodes (paragraph, heading, blockquote, list, table, codeBlock, mediaGroup)
- Handles inline marks (strong, em, strike, code, link, textColor)
- Converts ADF to markdown for terminal display
- `ReplaceAll()` for text substitution within ADF trees

### 7.2 Markdown Conversion (`pkg/md/`)

- `ToJiraMD()` - CommonMark to Jira wiki markup (for v2 API)
- `FromJiraMD()` - Jira wiki markup to CommonMark (for display)
- Uses `blackfriday` parser with `blackfriday-confluence` renderer

---

## 8. Query System

### 8.1 JQL Builder (`pkg/jql/jql.go`)

Fluent API for constructing Jira Query Language strings:

```go
q := jql.NewJQL("PROJ")           // project="PROJ"
q.FilterBy("type", "Bug")          // type="Bug"
q.In("status", "Open", "Closed")   // status IN ("Open", "Closed")
q.Gt("createdDate", "2024-01-01", true)  // createdDate>"2024-01-01"
q.OrderBy("created", "DESC")
```

Special values: `x` = IS EMPTY, `~value` = NOT EQUAL, `~x` = IS NOT EMPTY

### 8.2 Query Translation (`internal/query/issue.go`)

Translates CLI flags to JQL:
- Date filters support: `today`, `week`, `month`, `year`, `yyyy-mm-dd`, period format (`-10d`)
- Labels/status support positive and negative (`~label` = exclude)
- Pagination: `<from>:<limit>` format, max 100 per request
- Auto-detects order-by field based on which date filters are active

---

## 9. View/Rendering System (`internal/view/`)

Multiple output modes:
- **Interactive TUI**: tview-based table with keyboard navigation, issue preview pane
- **Plain text**: Tab-delimited table with optional headers
- **CSV**: Comma-separated output
- **Raw JSON**: Direct JSON dump
- **Markdown**: glamour-rendered markdown for issue details

Configurable columns, custom delimiters, timezone-aware date formatting.

---

## 10. Key Dependencies

| Package | Purpose |
|---------|---------|
| `spf13/cobra` | CLI framework |
| `spf13/viper` | Configuration management |
| `AlecAivazis/survey/v2` | Interactive prompts |
| `rivo/tview` | Terminal UI widgets |
| `gdamore/tcell/v2` | Terminal control |
| `charmbracelet/glamour` | Markdown rendering |
| `zalando/go-keyring` | Secure credential storage |
| `russross/blackfriday/v2` | Markdown parsing |
| `kentaro-m/blackfriday-confluence` | Jira wiki rendering |
| `pkg/browser` | System browser launch |

---

## 11. Build System (`Makefile`)

- `make build` - Vendor deps + build with ldflags (version, git commit, source date epoch)
- `make install` - Install binary
- `make lint` - golangci-lint (auto-installs if missing)
- `make test` - Race-enabled tests
- `make ci` - lint + test
- `make jira.server` - Docker Compose for local Jira instance
- CGO disabled by default for static binaries

---

## 12. Design Takeaways for gerrit-cli

Key patterns worth adopting:

1. **Proxy layer for API versioning** - Clean abstraction when supporting multiple API versions
2. **Singleton client with cascading config** - Flags > env > config file > netrc > keyring
3. **JQL-style query builder** - Fluent API for constructing Gerrit query strings
4. **Dual output modes** - Interactive TUI for exploration, plain/JSON for scripting
5. **Custom JSON marshalers** - Handle dynamic fields without reflection
6. **Interactive config wizard** - First-run setup with validation and server verification
7. **Consistent API method pattern** - Every operation follows: build path, marshal, call, check status, decode
8. **Add/remove semantics with `-` prefix** - Intuitive for editing list fields (labels, reviewers, etc.)
