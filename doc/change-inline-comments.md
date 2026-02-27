# Add Inline Comment Support (Draft Workflow)

## Context
The `gerrit change review` command can post messages and label votes, but has no inline comment support. Users want to add inline comments to specific file lines, review them before publishing, then publish all at once with a review.

## Design
Support a draft-based workflow: save inline comments as drafts, list/review them, then publish all drafts when posting a review.

The API plumbing already exists in `pkg/gerrit/comment.go`: `CreateDraft`, `GetDrafts`, `DeleteDraft`. The `SetReview` endpoint automatically publishes any pending drafts for the revision.

### Workflow
```bash
# 1. Save comments as drafts (not visible to others yet)
gerrit change draft <change-id> -c "main.go:42:fix this"
gerrit change draft <change-id> -c "pkg/util.go:10:typo here"

# 2. Review your drafts
gerrit change drafts <change-id>

# 3. Publish all drafts with a review (drafts become visible)
gerrit change review <change-id> -m "see inline comments"

# Or publish with inline comments directly (skip draft step)
gerrit change review <change-id> --comment "main.go:42:fix this"
```

## Changes

### 1. `internal/cmd/change/draft/draft.go` (new)
- New `gerrit change draft <change-id>` command
- `-c` / `--comment` flag (repeatable): format `file:line:message`
- `-p` / `--patchset` flag: target patchset (default: current)
- Parse each comment, call `client.CreateDraft()` for each
- Print confirmation with draft count

### 2. `internal/cmd/change/drafts/drafts.go` (new)
- New `gerrit change drafts <change-id>` command
- `-p` / `--patchset` flag: target patchset (default: current)
- Call `client.GetDrafts()`, render file:line + message for each

### 3. `internal/cmd/change/review/review.go` (modify)
- Add `comments []string` to the var block
- Add `--comment` / `-c` flag: `StringArrayVarP(&comments, "comment", "c", nil, "Inline comment (file:line:message)")`
- Parse each entry: `SplitN(s, ":", 3)`, convert line to int, validate 3 parts
- Build `map[string][]gerrit.CommentInput` and assign to `input.Comments`
- Exit with error if any comment has invalid format
- Note: `SetReview` also publishes any existing drafts for the revision automatically

### 4. `internal/cmd/change/change.go` (modify)
- Wire `draft` and `drafts` subcommands

### No API layer changes needed
`CreateDraft`, `GetDrafts`, `DeleteDraft` already exist in `pkg/gerrit/comment.go`. `ReviewInput.Comments` and `CommentInput` already exist in `pkg/gerrit/types.go`.

## TODO

### Phase 1 — Comment parsing helper

- [x] 1.1 Add `parseComment(s string) (file, line, message, error)` helper in `review.go`
  - `SplitN(s, ":", 3)`, validate 3 parts, convert line to int
  - Return clear error on bad format: `invalid comment format "...": expected file:line:message`

### Phase 2 — Draft commands

- [x] 2.1 Create `internal/cmd/change/draft/draft.go`
  - Cobra command `draft`, takes `<change-id>` as arg
  - Flags: `-c`/`--comment` (stringArray, required), `-p`/`--patchset` (int, default 0 → "current")
  - For each `-c` value: parse with `parseComment`, call `client.CreateDraft()` with `DraftInput{Path, Line, Message}`
  - Print `"Saved N draft comment(s)."` on success

- [x] 2.2 Create `internal/cmd/change/drafts/drafts.go`
  - Cobra command `drafts`, takes `<change-id>` as arg
  - Flags: `-p`/`--patchset` (int, default 0 → "current")
  - Call `client.GetDrafts()`, iterate map entries
  - Render each draft as `file:line — message` (one per line)
  - Print `"No drafts."` if empty

- [x] 2.3 Wire draft commands in `internal/cmd/change/change.go`
  - Import `draft` and `drafts` packages
  - Add both to `cmd.AddCommand()`

### Phase 3 — Inline comments on review

- [x] 3.1 Update `internal/cmd/change/review/review.go`
  - Add `comments []string` var
  - Add `-c`/`--comment` flag (stringArray)
  - Parse each with `parseComment`, build `map[string][]gerrit.CommentInput`
  - Assign to `input.Comments`
  - Exit with error on invalid format

### Phase 4 — Build & verify

- [x] 4.1 `go build ./...` compiles clean
- [ ] 4.2 `gerrit change draft <id> -c "file:line:msg"` saves draft
- [ ] 4.3 `gerrit change drafts <id>` lists drafts
- [ ] 4.4 `gerrit change review <id> -m "done"` publishes pending drafts
- [ ] 4.5 `gerrit change review <id> --comment "file:line:msg"` direct publish
- [ ] 4.6 `--comment "badformat"` prints parse error
