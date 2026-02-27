# gerrit-cli: Code Review CLI — Implementation Plan

## Context

We're building a new Gerrit CLI tool focused on code review workflows. Two reference codebases inform the design:
- `reference/jira-cli/` — clean layered architecture with typed structs, proxy layer, fluent query builder, separate view layer
- `reference/gerrit-cli/` — working Gerrit CLI but uses `map[string]interface{}` everywhere, flat command structure, mixed SSH/REST

The new project takes the best architectural patterns from jira-cli and applies them to Gerrit's REST API, producing a cleaner, more maintainable tool.

**Decisions:** Fresh Go project, REST API only (no SSH), scope = core review + diff viewing.

---

## Project Structure

```
gerrit-cli/
├── cmd/gerrit/main.go                    # Entry point
├── api/client.go                         # Singleton factory, viper config cascade
├── pkg/
│   ├── gerrit/
│   │   ├── client.go                     # HTTP client (auth, XSS strip, functional opts)
│   │   ├── types.go                      # All Gerrit API typed structs
│   │   ├── change.go                     # List, Get, Submit, Abandon, Rebase
│   │   ├── comment.go                    # GetComments, GetDrafts, CreateDraft
│   │   ├── review.go                     # SetReview (vote + message + inline comments)
│   │   ├── diff.go                       # GetFileDiff
│   │   ├── patch.go                      # GetPatch (download patch file)
│   │   └── reviewer.go                   # Add/Remove reviewer
│   └── query/
│       └── query.go                      # Gerrit search query builder (fluent)
├── internal/
│   ├── cmd/
│   │   ├── root/root.go                  # Root command, persistent flags, viper init
│   │   ├── init/init.go                  # Interactive config wizard
│   │   └── change/
│   │       ├── change.go                 # Parent "change" command
│   │       ├── list/list.go              # List changes with filters
│   │       ├── view/view.go              # View change details
│   │       ├── comments/comments.go      # View threaded comments
│   │       ├── review/review.go          # Vote + post message + inline comments
│   │       ├── submit/submit.go          # Submit/merge
│   │       ├── abandon/abandon.go        # Abandon
│   │       ├── rebase/rebase.go          # Server-side rebase
│   │       ├── diff/diff.go              # View file diffs
│   │       ├── patch/patch.go            # Download patch file
│   │       └── reviewers/
│   │           ├── reviewers.go          # Parent "reviewers" command
│   │           ├── add/add.go            # Add reviewer/CC
│   │           └── remove/remove.go      # Remove reviewer
│   ├── view/
│   │   ├── change.go                     # Single change detail + patchset history rendering
│   │   ├── changes.go                    # Change list table rendering
│   │   ├── comments.go                   # Threaded comment rendering (grouped by patchset)
│   │   └── diff.go                       # Unified diff with ANSI colors
│   ├── config/
│   │   └── config.go                     # Config struct, Load/Save, env overrides
│   └── cmdutil/
│       └── cmdutil.go                    # ExitIfError, FormatTimeAgo, color helpers
├── go.mod
├── go.sum
└── Makefile
```

## Command Hierarchy

```
gerrit
├── init                                  # Setup config wizard
├── change
│   ├── list (ls)                         # List changes with filters
│   ├── view <id>                         # View change details
│   ├── comments <id>                     # View threaded comments
│   ├── review <id>                       # Vote + post message
│   ├── submit <id>                       # Submit/merge
│   ├── abandon <id>                      # Abandon
│   ├── rebase <id>                       # Server-side rebase
│   ├── diff <id>                         # View file diffs
│   ├── patch <id>                        # Download patch file
│   └── reviewers
│       ├── add <id> <account>            # Add reviewer/CC
│       └── remove <id> <account>         # Remove reviewer
└── version
```

---

## Gerrit REST API Endpoints

All endpoints use `/a/` prefix for authenticated access. Basic auth via `Authorization: Basic base64(user:password)`. All JSON responses prefixed with `)]}'` (XSS protection) — stripped in client.

The `{revision}` parameter accepts a patchset number (e.g. `3`), a commit SHA, or the literal `"current"` for the latest patchset. All client methods default to `"current"` when revision is empty.

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List changes | GET | `/changes/?q={query}&n={limit}&o=LABELS&o=CURRENT_REVISION&o=DETAILED_ACCOUNTS` |
| Get change | GET | `/changes/{id}?o=LABELS&o=CURRENT_REVISION&o=CURRENT_COMMIT&o=DETAILED_ACCOUNTS&o=DETAILED_LABELS` |
| Get change (all patchsets) | GET | `/changes/{id}?o=ALL_REVISIONS&o=ALL_COMMITS&o=DETAILED_ACCOUNTS&o=DETAILED_LABELS` |
| Get comments | GET | `/changes/{id}/comments` |
| Get files | GET | `/changes/{id}/revisions/{revision}/files` |
| Get diff | GET | `/changes/{id}/revisions/{revision}/files/{file}/diff` |
| Get diff (inter-patchset) | GET | `/changes/{id}/revisions/{revision}/files/{file}/diff?base={base_patchset}` |
| Post review | POST | `/changes/{id}/revisions/{revision}/review` |
| Submit | POST | `/changes/{id}/submit` |
| Abandon | POST | `/changes/{id}/abandon` |
| Rebase | POST | `/changes/{id}/rebase` |
| Add reviewer | POST | `/changes/{id}/reviewers` |
| Remove reviewer | DELETE | `/changes/{id}/reviewers/{account}` |
| Get patch | GET | `/changes/{id}/revisions/{revision}/patch` |
| Create draft comment | PUT | `/changes/{id}/revisions/{revision}/drafts` |
| List draft comments | GET | `/changes/{id}/revisions/{revision}/drafts` |
| Delete draft comment | DELETE | `/changes/{id}/revisions/{revision}/drafts/{draft_id}` |
| Server version | GET | `/config/server/version` |

Note: Inline comments are posted via the `comments` field in the Post review (`SetReview`) request body. Draft comments allow composing inline comments before publishing them as part of a review.

Note: The Get patch endpoint returns base64-encoded patch content. Add `?zip` for zip format or `?download` for raw download.



## Key Types (`pkg/gerrit/types.go`)

Full Go struct definitions for all Gerrit API responses:

```go
package gerrit

// AccountInfo represents a Gerrit account.
type AccountInfo struct {
    AccountID int    `json:"_account_id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    Username  string `json:"username"`
}

// DisplayName returns the best available display name.
func (a *AccountInfo) DisplayName() string {
    if a.Name != "" { return a.Name }
    if a.Username != "" { return a.Username }
    if a.Email != "" { return a.Email }
    return "unknown"
}

// LabelInfo represents a single label on a change.
type LabelInfo struct {
    Approved     *AccountInfo      `json:"approved,omitempty"`
    Rejected     *AccountInfo      `json:"rejected,omitempty"`
    Recommended  *AccountInfo      `json:"recommended,omitempty"`
    Disliked     *AccountInfo      `json:"disliked,omitempty"`
    All          []ApprovalInfo    `json:"all,omitempty"`
    Values       map[string]string `json:"values,omitempty"`
    DefaultValue int               `json:"default_value"`
}

// ApprovalInfo represents a single vote on a label.
type ApprovalInfo struct {
    AccountInfo
    Value int    `json:"value"`
    Date  string `json:"date"`
}

// RevisionInfo represents a single patchset revision.
type RevisionInfo struct {
    Number  int        `json:"_number"`
    Ref     string     `json:"ref"`
    Created string     `json:"created"`
    Commit  CommitInfo `json:"commit"`
}

// CommitInfo represents a commit.
type CommitInfo struct {
    Subject   string        `json:"subject"`
    Message   string        `json:"message"`
    Author    GitPersonInfo `json:"author"`
    Committer GitPersonInfo `json:"committer"`
    Parents   []ParentInfo  `json:"parents"`
}

// GitPersonInfo represents a git author/committer.
type GitPersonInfo struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Date  string `json:"date"`
}

// ParentInfo represents a parent commit.
type ParentInfo struct {
    Commit  string `json:"commit"`
    Subject string `json:"subject"`
}

// ChangeInfo represents a Gerrit change.
type ChangeInfo struct {
    ID                     string                   `json:"id"`
    Project                string                   `json:"project"`
    Branch                 string                   `json:"branch"`
    Topic                  string                   `json:"topic,omitempty"`
    ChangeID               string                   `json:"change_id"`
    Subject                string                   `json:"subject"`
    Status                 string                   `json:"status"`
    Created                string                   `json:"created"`
    Updated                string                   `json:"updated"`
    Submitted              string                   `json:"submitted,omitempty"`
    Submitter              *AccountInfo             `json:"submitter,omitempty"`
    Owner                  AccountInfo              `json:"owner"`
    Number                 int                      `json:"_number"`
    Labels                 map[string]LabelInfo     `json:"labels,omitempty"`
    Reviewers              map[string][]AccountInfo `json:"reviewers,omitempty"`
    CurrentRevision        string                   `json:"current_revision,omitempty"`
    Revisions              map[string]RevisionInfo  `json:"revisions,omitempty"`
    Mergeable              bool                     `json:"mergeable,omitempty"`
    Insertions             int                      `json:"insertions"`
    Deletions              int                      `json:"deletions"`
    TotalCommentCount      int                      `json:"total_comment_count"`
    UnresolvedCommentCount int                      `json:"unresolved_comment_count"`
    MoreChanges            bool                     `json:"_more_changes,omitempty"`
}

// FileInfo represents a changed file in a revision.
type FileInfo struct {
    Status        string `json:"status"`
    LinesInserted int    `json:"lines_inserted,omitempty"`
    LinesDeleted  int    `json:"lines_deleted,omitempty"`
    SizeDelta     int    `json:"size_delta"`
    Size          int    `json:"size"`
    OldPath       string `json:"old_path,omitempty"`
}

// DiffInfo represents a file diff.
type DiffInfo struct {
    MetaA      DiffFileMetaInfo `json:"meta_a,omitempty"`
    MetaB      DiffFileMetaInfo `json:"meta_b,omitempty"`
    ChangeType string           `json:"change_type"`
    Content    []DiffContent    `json:"content"`
}

// DiffFileMetaInfo represents metadata about a file in a diff.
type DiffFileMetaInfo struct {
    Name        string `json:"name"`
    ContentType string `json:"content_type"`
    Lines       int    `json:"lines"`
}

// DiffContent represents a section of diff content.
type DiffContent struct {
    A    []string `json:"a,omitempty"`    // Lines only in side A (removed)
    B    []string `json:"b,omitempty"`    // Lines only in side B (added)
    AB   []string `json:"ab,omitempty"`   // Lines in both (context)
    Skip int      `json:"skip,omitempty"` // Number of lines skipped
}

// CommentInfo represents a review comment.
type CommentInfo struct {
    ID         string       `json:"id"`
    Path       string       `json:"path,omitempty"`
    Line       int          `json:"line,omitempty"`
    Range      *CommentRange `json:"range,omitempty"`
    InReplyTo  string       `json:"in_reply_to,omitempty"`
    Message    string       `json:"message"`
    Updated    string       `json:"updated"`
    Author     AccountInfo  `json:"author"`
    Unresolved bool         `json:"unresolved"`
    PatchSet   int          `json:"patch_set,omitempty"`
}

// CommentRange represents a range within a file.
type CommentRange struct {
    StartLine      int `json:"start_line"`
    StartCharacter int `json:"start_character"`
    EndLine        int `json:"end_line"`
    EndCharacter   int `json:"end_character"`
}

// ChangeMessageInfo represents a change message/history entry.
type ChangeMessageInfo struct {
    ID             string      `json:"id"`
    Author         AccountInfo `json:"author"`
    RealAuthor     AccountInfo `json:"real_author,omitempty"`
    Date           string      `json:"date"`
    Message        string      `json:"message"`
    RevisionNumber int         `json:"_revision_number"`
}

// --- Request types ---

// ReviewInput is the request body for posting a review.
type ReviewInput struct {
    Message  string                      `json:"message,omitempty"`
    Labels   map[string]int              `json:"labels,omitempty"`
    Comments map[string][]CommentInput   `json:"comments,omitempty"`
}

// CommentInput is a comment to post as part of a review.
type CommentInput struct {
    Line       int    `json:"line,omitempty"`
    Message    string `json:"message"`
    Unresolved bool   `json:"unresolved,omitempty"`
}

// DraftInput is the request body for creating/updating a draft comment.
type DraftInput struct {
    Path       string        `json:"path"`
    Line       int           `json:"line,omitempty"`
    Range      *CommentRange `json:"range,omitempty"`
    InReplyTo  string        `json:"in_reply_to,omitempty"`
    Message    string        `json:"message"`
    Unresolved bool          `json:"unresolved,omitempty"`
}

// ReviewerInput is the request body for adding a reviewer.
type ReviewerInput struct {
    Reviewer string `json:"reviewer"`
    State    string `json:"state,omitempty"` // "REVIEWER" or "CC"
}

// RebaseInput is the request body for rebasing a change.
type RebaseInput struct {
    Base           string `json:"base,omitempty"`
    AllowConflicts bool   `json:"allow_conflicts,omitempty"`
}

// AbandonInput is the request body for abandoning a change.
type AbandonInput struct {
    Message string `json:"message,omitempty"`
}

// SubmitInput is the request body for submitting a change.
type SubmitInput struct{}

// AddReviewerResult is the response from adding a reviewer.
type AddReviewerResult struct {
    Input     string        `json:"input"`
    Reviewers []AccountInfo `json:"reviewers,omitempty"`
    CCs       []AccountInfo `json:"ccs,omitempty"`
    Error     string        `json:"error,omitempty"`
}
```

---

## Client Design (`pkg/gerrit/client.go`)

Pattern from: `reference/jira-cli/pkg/jira/client.go`

```go
// Config holds Gerrit connection configuration.
type Config struct {
    Server   string
    Username string
    Password string // HTTP password/token
    Insecure bool
    Debug    bool
}

// ClientFunc decorates option for client.
type ClientFunc func(*Client)

// Client is the Gerrit REST API client.
type Client struct {
    server    string
    username  string
    password  string
    insecure  bool
    debug     bool
    timeout   time.Duration
    transport http.RoundTripper
}

func NewClient(cfg Config, opts ...ClientFunc) *Client
func WithTimeout(d time.Duration) ClientFunc
func WithInsecureTLS(b bool) ClientFunc

// Low-level (handle auth header + XSS prefix strip + error formatting)
func (c *Client) Get(ctx context.Context, path string) ([]byte, error)
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error)
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error)
func (c *Client) Delete(ctx context.Context, path string) error
```

Key details:
- URL format: `{server}/a/{path}` (the `/a/` prefix enables authenticated access)
- Auth: `req.SetBasicAuth(username, password)`
- XSS strip: `bytes.TrimPrefix(body, []byte(")]}'"))` then trim leading newline
- Error type: `ErrUnexpectedResponse { Status string; StatusCode int; Body string }`
- Debug mode: dump request/response via `httputil`

---

## API Client Method Signatures

### `pkg/gerrit/change.go`
```go
func (c *Client) ListChanges(ctx context.Context, query string, limit int, opts ...string) ([]ChangeInfo, error)
func (c *Client) GetChange(ctx context.Context, changeID string, opts ...string) (*ChangeInfo, error)
// opts includes standard Gerrit query options: "CURRENT_REVISION", "ALL_REVISIONS",
// "CURRENT_COMMIT", "ALL_COMMITS", "DETAILED_LABELS", "DETAILED_ACCOUNTS", etc.
func (c *Client) GetChangeFiles(ctx context.Context, changeID, revision string) (map[string]FileInfo, error)
func (c *Client) GetChangeMessages(ctx context.Context, changeID string) ([]ChangeMessageInfo, error)
func (c *Client) Submit(ctx context.Context, changeID string) (*ChangeInfo, error)
func (c *Client) Abandon(ctx context.Context, changeID string, input *AbandonInput) (*ChangeInfo, error)
func (c *Client) Restore(ctx context.Context, changeID string) (*ChangeInfo, error)
func (c *Client) Rebase(ctx context.Context, changeID string, input *RebaseInput) (*ChangeInfo, error)
```

### `pkg/gerrit/comment.go`
```go
func (c *Client) GetComments(ctx context.Context, changeID string) (map[string][]CommentInfo, error)
func (c *Client) GetDrafts(ctx context.Context, changeID, revision string) (map[string][]CommentInfo, error)
func (c *Client) CreateDraft(ctx context.Context, changeID, revision string, input *DraftInput) (*CommentInfo, error)
func (c *Client) DeleteDraft(ctx context.Context, changeID, revision, draftID string) error
```

### `pkg/gerrit/review.go`
```go
func (c *Client) SetReview(ctx context.Context, changeID, revision string, input *ReviewInput) error
```

### `pkg/gerrit/diff.go`
```go
// GetFileDiff gets the diff for a file. base=0 means diff against parent;
// base>0 means diff against that patchset number (inter-patchset diff).
func (c *Client) GetFileDiff(ctx context.Context, changeID, revision, file string, base int) (*DiffInfo, error)
```

### `pkg/gerrit/patch.go`
```go
func (c *Client) GetPatch(ctx context.Context, changeID, revision string) ([]byte, error)
```

### `pkg/gerrit/reviewer.go`
```go
func (c *Client) AddReviewer(ctx context.Context, changeID string, input *ReviewerInput) (*AddReviewerResult, error)
func (c *Client) RemoveReviewer(ctx context.Context, changeID, accountID string) error
```

---

## API Proxy Layer (`api/client.go`)

Singleton factory with viper config cascade (pattern from `reference/jira-cli/api/client.go`):

```go
package api

var gerritClient *gerrit.Client

func Client(cfg gerrit.Config) *gerrit.Client {
    if gerritClient != nil {
        return gerritClient
    }
    // Cascade: explicit config -> viper -> env vars
    if cfg.Server == "" {
        cfg.Server = viper.GetString("server")
    }
    if cfg.Username == "" {
        cfg.Username = viper.GetString("username")
    }
    if cfg.Password == "" {
        cfg.Password = viper.GetString("password")
    }
    if cfg.Password == "" {
        cfg.Password = os.Getenv("GERRIT_HTTP_PASSWORD")
    }
    gerritClient = gerrit.NewClient(cfg, gerrit.WithTimeout(15*time.Second))
    return gerritClient
}

func DefaultClient(debug bool) *gerrit.Client {
    return Client(gerrit.Config{Debug: debug})
}
```

---

## Query Builder (`pkg/query/query.go`)

Gerrit search uses space-separated `key:value` predicates.

```go
query.New().Owner("self").Status("open").Project("myproject").Limit(25).String()
// → "owner:self status:open project:myproject"

// Methods: Owner, Reviewer, Status, Project, Branch, Topic, Label, After, Before,
//          Is, Has, Raw, Limit, String
```

---

## Patchset Resolution

Gerrit changes contain multiple patchsets (revisions). Commands that operate on a specific patchset accept `--patchset/-p <N>` where N is the patchset number. When omitted, defaults to the current (latest) patchset.

Resolution logic in commands:
1. If `--patchset` is given, convert patchset number to revision ID by fetching the change with `ALL_REVISIONS` and looking up the matching `RevisionInfo._number`
2. If omitted, use the literal string `"current"` in the API path

Commands supporting `--patchset`: `diff`, `comments`, `review`, `patch`

The `change view --patchsets` flag fetches the change with `ALL_REVISIONS` + `ALL_COMMITS` and renders a patchset history table showing: PS number, commit SHA (short), author, created date, and commit subject.

---

## Command Flags

| Command | Flags |
|---------|-------|
| `change list` | `--status/-s` (open), `--reviewer` (bool), `--project/-p`, `--branch/-b`, `--query/-q` (raw), `--limit/-n` (25), `--plain`, `--no-headers`, `--json` |
| `change view` | `--patchsets` (bool, show all patchset history), `--files` (bool), `--json` |
| `change comments` | `--patchset/-p` (int, filter by patchset), `--all` (bool, default unresolved only), `--json` |
| `change review` | `--patchset/-p` (int, default current), `--message/-m`, `--code-review` (int), `--verified` (int), `--label` (stringArray "Name=score") |
| `change submit` | (none) |
| `change abandon` | `--message/-m` |
| `change rebase` | `--base/-b`, `--allow-conflicts` |
| `change diff` | `--patchset/-p` (int, default current), `--base` (int, compare against patchset N), `--file/-f` (specific file), `--context/-C` (3), `--no-color`, `--stat` (diffstat only) |
| `change patch` | `--patchset/-p` (int, default current), `--output/-o` (file path, default stdout), `--zip` (bool) |
| `change reviewers add` | `--cc` (bool, add as CC) |

---

## Diff Rendering (`internal/view/diff.go`)

Gerrit returns structured `DiffInfo` with `Content []DiffContent` sections:
- `AB []string` — context lines (both sides)
- `A []string` — removed lines (old side only)
- `B []string` — added lines (new side only)
- `Skip int` — skipped lines count

Rendering strategy:
1. Print file header: `--- a/{old_path}` / `+++ b/{new_path}`
2. Track `lineA` and `lineB` counters
3. Generate hunk headers `@@ -lineA,count +lineB,count @@` at boundaries
4. AB → dim/default, A → red with `- ` prefix, B → green with `+ ` prefix
5. Skip → `... N lines skipped ...`
6. `--stat` mode: file list with `+N -M` counts (like `git diff --stat`)
7. `--context` trims AB sections to N lines around changes

---

## Config (`internal/config/config.go`)

Location: `~/.gerrit/.config.yml`

```yaml
server: https://gerrit.example.com
username: john.doe
project: my-project          # optional default
```

Password via `GERRIT_HTTP_PASSWORD` env var (recommended, never stored in file).
Env overrides: `GERRIT_SERVER`, `GERRIT_USERNAME`, `GERRIT_PROJECT`.

Init wizard: prompt server → username → test connection via `GET /config/server/version` → optionally select default project → save.

---

## Implementation Order

### Phase 1 — Foundation
1. `go.mod` — cobra, viper, fatih/color, survey/v2
2. `pkg/gerrit/types.go` — all typed structs
3. `pkg/gerrit/client.go` — HTTP client with auth, XSS strip, functional options
4. `internal/config/config.go` — config struct, Load/Save, viper integration
5. `internal/cmdutil/cmdutil.go` — ExitIfError, FormatTimeAgo, FormatScore, color helpers
6. `cmd/gerrit/main.go` + `internal/cmd/root/root.go` — entry point + root command

### Phase 2 — API Operations
7. `pkg/gerrit/change.go` — ListChanges, GetChange, GetChangeFiles, GetChangeMessages, Submit, Abandon, Rebase
8. `pkg/gerrit/comment.go` — GetComments, GetDrafts, CreateDraft, DeleteDraft
9. `pkg/gerrit/review.go` — SetReview
10. `pkg/gerrit/diff.go` — GetFileDiff
11. `pkg/gerrit/patch.go` — GetPatch
12. `pkg/gerrit/reviewer.go` — AddReviewer, RemoveReviewer
13. `api/client.go` — singleton factory with viper cascade
14. `pkg/query/query.go` — fluent query builder

### Phase 3 — Commands + Views
15. `internal/cmd/init/init.go` — interactive config wizard
16. `internal/view/changes.go` + `internal/cmd/change/list/list.go` — list changes
17. `internal/view/change.go` + `internal/cmd/change/view/view.go` — view change detail
18. `internal/view/comments.go` + `internal/cmd/change/comments/comments.go` — threaded comments
19. `internal/cmd/change/review/review.go` — vote + message + inline comments
20. `internal/cmd/change/submit/submit.go` — submit
21. `internal/cmd/change/abandon/abandon.go` — abandon
22. `internal/cmd/change/rebase/rebase.go` — rebase
23. `internal/cmd/change/reviewers/` — add + remove reviewers
24. `internal/view/diff.go` + `internal/cmd/change/diff/diff.go` — diff viewing
25. `internal/cmd/change/patch/patch.go` — download patch file
26. `internal/cmd/change/change.go` — wire all subcommands
27. `Makefile` — build, clean, install targets

---

## Verification

1. `go build ./...` — compiles without errors
2. `gerrit init` — creates config, tests connection to a Gerrit instance
3. `gerrit change list` — returns formatted change list
4. `gerrit change view <id>` — shows change details with labels and reviewers
5. `gerrit change comments <id>` — shows threaded comments
6. `gerrit change review <id> --code-review +1 -m "LGTM"` — posts review
7. `gerrit change diff <id>` — shows colored unified diff
8. `gerrit change diff <id> --stat` — shows diffstat summary
9. `gerrit change diff <id> --patchset 3` — diff for patchset 3
10. `gerrit change diff <id> --patchset 3 --base 1` — inter-patchset diff (PS1 → PS3)
11. `gerrit change view <id> --patchsets` — shows patchset history
12. `gerrit change comments <id> --patchset 2` — comments on patchset 2
13. `gerrit change patch <id>` — downloads patch to stdout
14. `gerrit change patch <id> -o patch.diff` — saves patch to file

---

## Key Reference Files

| File | What to borrow |
|------|---------------|
| `reference/jira-cli/pkg/jira/client.go` | HTTP client pattern: functional options, auth, request/response flow |
| `reference/jira-cli/api/client.go` | Singleton factory with viper config cascade |
| `reference/jira-cli/pkg/jql/jql.go` | Fluent query builder pattern |
| `reference/jira-cli/internal/cmd/root/root.go` | Root command + viper init |
| `reference/jira-cli/internal/cmd/issue/issue.go` | Subcommand hierarchy wiring |
| `reference/gerrit-cli/internal/gerrit/rest.go` | Gerrit-specific: XSS prefix, `/a/` prefix, basic auth |
| `reference/gerrit-cli/internal/utils/format.go` | Color helpers, table rendering, time formatting, score display |

---

## TODO

### Phase 1 — Foundation

- [x] 1.1 Initialize Go module
  - `go mod init` with module path
  - Add dependencies: cobra, viper, fatih/color, survey/v2
  - Run `go mod tidy`

- [x] 1.2 Create `pkg/gerrit/types.go`
  - All response structs: AccountInfo, LabelInfo, ApprovalInfo, RevisionInfo, CommitInfo, GitPersonInfo, ParentInfo, ChangeInfo, FileInfo, DiffInfo, DiffFileMetaInfo, DiffContent, CommentInfo, CommentRange, ChangeMessageInfo
  - All request structs: ReviewInput, CommentInput, DraftInput, ReviewerInput, RebaseInput, AbandonInput, SubmitInput
  - Response struct: AddReviewerResult
  - Helper method: AccountInfo.DisplayName()

- [x] 1.3 Create `pkg/gerrit/client.go`
  - Config struct (Server, Username, Password, Insecure, Debug, NoAuthPrefix)
  - Client struct with unexported fields
  - ClientFunc type + NewClient constructor with functional options
  - WithTimeout, WithInsecureTLS option functions
  - ErrUnexpectedResponse error type (Status, StatusCode, Body)
  - Internal helper: buildURL — `{server}/a/{path}` (or `{server}/{path}` when NoAuthPrefix)
  - Internal helper: do — execute request, set basic auth header, read body, strip XSS prefix `)]}'`, check status code
  - Public methods: Get, Post, Put, Delete
  - Debug mode: dump request/response via httputil when enabled
  - TLS: skip verify when Insecure is true

- [x] 1.4 Create `internal/config/config.go`
  - Config struct: Server, Username, Project (optional default)
  - ConfigDir: `~/.gerrit/`
  - ConfigFile: `~/.gerrit/.config.yml`
  - Load() — read YAML, unmarshal into Config
  - Save() — marshal Config, write YAML
  - Viper integration: bind config keys, set config path
  - Env overrides: GERRIT_SERVER, GERRIT_USERNAME, GERRIT_PROJECT

- [x] 1.5 Create `internal/cmdutil/cmdutil.go`
  - ExitIfError(err) — print error + os.Exit(1)
  - FormatTimeAgo(timestamp string) — parse Gerrit timestamp, return relative time
  - FormatScore(value int) — "+1", "+2", "-1", "-2" with color
  - Color helpers: Bold, Dim, Red, Green, Yellow, Cyan wrappers using fatih/color
  - PatchsetRevision(change *ChangeInfo, patchset int) — resolve patchset number to revision ID

- [x] 1.6 Create `cmd/gerrit/main.go` + `internal/cmd/root/root.go`
  - main.go: call root.Execute()
  - root.go: cobra root command "gerrit" with description
  - PersistentPreRun: init viper, read config file, bind env vars
  - Persistent flags: --debug, --config (override config path)
  - Version subcommand: print build version
  - Wire init command (placeholder until Phase 3)

- [x] 1.7 Verify Phase 1: `go build ./...` compiles clean

### Phase 2 — API Operations

- [x] 2.1 Create `pkg/gerrit/change.go`
  - ListChanges(ctx, query, limit, opts) — GET /changes/?q=...&n=...&o=...
  - GetChange(ctx, changeID, opts) — GET /changes/{id}?o=... (supports ALL_REVISIONS, ALL_COMMITS, etc.)
  - GetChangeFiles(ctx, changeID, revision) — GET /changes/{id}/revisions/{revision}/files
  - GetChangeMessages(ctx, changeID) — GET /changes/{id}/messages (not in endpoint table, but used by view)
  - Submit(ctx, changeID) — POST /changes/{id}/submit
  - Abandon(ctx, changeID, input) — POST /changes/{id}/abandon
  - Restore(ctx, changeID) — POST /changes/{id}/restore
  - Rebase(ctx, changeID, input) — POST /changes/{id}/rebase

- [x] 2.2 Create `pkg/gerrit/comment.go`
  - GetComments(ctx, changeID) — GET /changes/{id}/comments
  - GetDrafts(ctx, changeID, revision) — GET /changes/{id}/revisions/{revision}/drafts
  - CreateDraft(ctx, changeID, revision, input) — PUT /changes/{id}/revisions/{revision}/drafts
  - DeleteDraft(ctx, changeID, revision, draftID) — DELETE /changes/{id}/revisions/{revision}/drafts/{id}

- [x] 2.3 Create `pkg/gerrit/review.go`
  - SetReview(ctx, changeID, revision, input) — POST /changes/{id}/revisions/{revision}/review
  - Input includes message, labels map, and inline comments map

- [x] 2.4 Create `pkg/gerrit/diff.go`
  - GetFileDiff(ctx, changeID, revision, file, base) — GET /changes/{id}/revisions/{revision}/files/{file}/diff
  - When base > 0, append ?base={base} for inter-patchset diff
  - URL-encode file path (slashes, special chars)

- [x] 2.5 Create `pkg/gerrit/patch.go`
  - GetPatch(ctx, changeID, revision) — GET /changes/{id}/revisions/{revision}/patch
  - Decode base64 response body into raw patch bytes

- [x] 2.6 Create `pkg/gerrit/reviewer.go`
  - AddReviewer(ctx, changeID, input) — POST /changes/{id}/reviewers
  - RemoveReviewer(ctx, changeID, accountID) — DELETE /changes/{id}/reviewers/{account}

- [x] 2.7 Create `api/client.go`
  - Package-level var gerritClient *gerrit.Client
  - Client(cfg) — singleton factory with viper config cascade (server, username, password from viper, then GERRIT_HTTP_PASSWORD env)
  - DefaultClient(debug) — shorthand using empty config + debug flag
  - NoAuthPrefix support for nginx-authed instances

- [x] 2.8 Create `pkg/query/query.go`
  - Query struct with internal string slice
  - New() constructor
  - Fluent methods: Owner, Reviewer, Status, Project, Branch, Topic, Label, After, Before, Is, Has, Raw
  - Limit stored separately (not part of query string, used by caller)
  - String() — join predicates with space

- [x] 2.9 Verify Phase 2: `go build ./...` compiles clean

### Phase 3 — Commands + Views

- [x] 3.1 Create `internal/cmd/init/init.go`
  - Interactive prompts: server URL, username (using survey/v2)
  - Test connection: GET /config/server/version
  - Optionally prompt for default project
  - Save config to ~/.gerrit/.config.yml
  - Print success message with config location

- [x] 3.2 Create `internal/view/changes.go`
  - RenderChangeList(changes []ChangeInfo, plain, noHeaders bool)
  - Table columns: Number, Subject (truncated), Owner, Project, Branch, Status, Updated, +/-, CR/V scores
  - Plain mode: tab-separated, no color
  - JSON mode handled at command level (json.Marshal + print)

- [x] 3.3 Create `internal/cmd/change/list/list.go`
  - Cobra command "list" with alias "ls"
  - Flags: --status, --reviewer, --project, --branch, --query, --limit, --plain, --no-headers, --json
  - Build query using pkg/query builder
  - If --project not given, use viper default project
  - Call ListChanges, pass to view.RenderChangeList or JSON output

- [x] 3.4 Create `internal/view/change.go`
  - RenderChangeDetail(change *ChangeInfo, showFiles bool, showPatchsets bool)
  - Sections: header (number + subject), status, owner, project/branch, created/updated, labels with votes, reviewers
  - --files: list changed files with status and +/- counts
  - --patchsets: table of all revisions — PS#, SHA (short), author, date, subject
  - Fetch with ALL_REVISIONS + ALL_COMMITS when --patchsets

- [x] 3.5 Create `internal/cmd/change/view/view.go`
  - Cobra command "view", takes change ID as arg
  - Flags: --patchsets, --files, --json
  - Call GetChange with appropriate opts, pass to view.RenderChangeDetail

- [x] 3.6 Create `internal/view/comments.go`
  - RenderComments(comments map[string][]CommentInfo, allComments bool, patchset int)
  - Group by file path, then thread by InReplyTo
  - Show: file:line, author, date, message, [UNRESOLVED] tag
  - Filter to unresolved only by default (--all shows all)
  - --patchset filters to comments on that patchset number

- [x] 3.7 Create `internal/cmd/change/comments/comments.go`
  - Cobra command "comments", takes change ID as arg
  - Flags: --patchset, --all, --json
  - Call GetComments, pass to view.RenderComments

- [x] 3.8 Create `internal/cmd/change/review/review.go`
  - Cobra command "review", takes change ID as arg
  - Flags: --patchset, --message, --code-review, --verified, --label
  - Build ReviewInput: parse --label "Name=score" pairs, set message
  - Resolve patchset to revision using cmdutil.PatchsetRevision
  - Call SetReview, print confirmation

- [x] 3.9 Create `internal/cmd/change/submit/submit.go`
  - Cobra command "submit", takes change ID as arg
  - Call Submit, print result status

- [x] 3.10 Create `internal/cmd/change/abandon/abandon.go`
  - Cobra command "abandon", takes change ID as arg
  - Flag: --message
  - Call Abandon with AbandonInput, print confirmation

- [x] 3.11 Create `internal/cmd/change/rebase/rebase.go`
  - Cobra command "rebase", takes change ID as arg
  - Flags: --base, --allow-conflicts
  - Call Rebase with RebaseInput, print result

- [x] 3.12 Create `internal/cmd/change/reviewers/`
  - reviewers.go: parent "reviewers" command
  - add/add.go: "add" command, takes change ID + account as args, flag --cc, call AddReviewer
  - remove/remove.go: "remove" command, takes change ID + account as args, call RemoveReviewer

- [x] 3.13 Create `internal/view/diff.go`
  - RenderDiff(files map[string]FileInfo, getDiff func, noColor bool, contextLines int)
  - RenderDiffStat(files map[string]FileInfo) — file list with +N -M bar graph
  - Per-file rendering: file header (--- a/ +++ b/), hunk headers (@@ ... @@), colored lines
  - Track lineA/lineB counters across DiffContent sections
  - Handle Skip sections: "... N lines skipped ..."
  - Context trimming: only show N lines of AB around A/B sections

- [x] 3.14 Create `internal/cmd/change/diff/diff.go`
  - Cobra command "diff", takes change ID as arg
  - Flags: --patchset, --base, --file, --context, --no-color, --stat
  - Resolve patchset to revision
  - If --stat: fetch files, render diffstat
  - If --file: fetch single file diff
  - Else: fetch all files, iterate and render each diff

- [x] 3.15 Create `internal/cmd/change/patch/patch.go`
  - Cobra command "patch", takes change ID as arg
  - Flags: --patchset, --output, --zip
  - Resolve patchset to revision
  - Call GetPatch, write to stdout or --output file

- [x] 3.16 Create `internal/cmd/change/change.go`
  - Parent "change" command
  - Wire all subcommands: list, view, comments, review, submit, abandon, rebase, diff, patch, reviewers

- [x] 3.17 Update `internal/cmd/root/root.go`
  - Wire change command and init command to root

- [x] 3.18 Create `Makefile`
  - `build`: go build -o bin/gerrit ./cmd/gerrit
  - `install`: go install ./cmd/gerrit
  - `clean`: rm -rf bin/

- [x] 3.19 Verify Phase 3: `go build ./...` compiles clean

### Phase 4 — Integration Verification

- [x] 4.1 `gerrit init` — creates config, tests connection
- [x] 4.2 `gerrit change list` — returns formatted change list
- [x] 4.3 `gerrit change view <id>` — shows change details with labels and reviewers
- [x] 4.4 `gerrit change view <id> --patchsets` — shows patchset history
- [x] 4.5 `gerrit change view <id> --files` — shows changed files
- [x] 4.6 `gerrit change comments <id>` — shows unresolved comments
- [x] 4.7 `gerrit change comments <id> --all` — shows all comments
- [x] 4.8 `gerrit change comments <id> --patchset 2` — comments on PS2
- [ ] 4.9 `gerrit change review <id> --code-review +1 -m "LGTM"` — posts review (blocked: server auth config)
- [x] 4.10 `gerrit change diff <id>` — shows colored unified diff
- [x] 4.11 `gerrit change diff <id> --stat` — shows diffstat summary
- [x] 4.12 `gerrit change diff <id> --patchset 3` — diff for specific patchset
- [x] 4.13 `gerrit change diff <id> --patchset 3 --base 1` — inter-patchset diff
- [x] 4.14 `gerrit change patch <id>` — downloads patch to stdout
- [x] 4.15 `gerrit change patch <id> -o patch.diff` — saves patch to file
- [ ] 4.16 `gerrit change submit <id>` — submits change (blocked: server auth config)
- [ ] 4.17 `gerrit change abandon <id> -m "reason"` — abandons change (blocked: server auth config)
- [ ] 4.18 `gerrit change rebase <id>` — rebases change (blocked: server auth config)
- [ ] 4.19 `gerrit change reviewers add <id> <account>` — adds reviewer (blocked: server auth config)
- [ ] 4.20 `gerrit change reviewers remove <id> <account>` — removes reviewer (blocked: server auth config)

