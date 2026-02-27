package view

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

type commentEntry struct {
	Path    string
	Comment gerrit.CommentInfo
}

func RenderComments(comments map[string][]gerrit.CommentInfo, allComments bool, patchset int) {
	var entries []commentEntry
	for path, cs := range comments {
		for _, c := range cs {
			if !allComments && !c.Unresolved {
				continue
			}
			if patchset > 0 && c.PatchSet != patchset {
				continue
			}
			entries = append(entries, commentEntry{path, c})
		}
	}
	if len(entries) == 0 {
		fmt.Println("No comments found.")
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		return entries[i].Comment.Updated < entries[j].Comment.Updated
	})

	// Build thread map
	threads := make(map[string][]commentEntry)
	var roots []commentEntry
	for _, e := range entries {
		if e.Comment.InReplyTo == "" {
			roots = append(roots, e)
		}
		threads[e.Comment.InReplyTo] = append(threads[e.Comment.InReplyTo], e)
	}

	currentPath := ""
	for _, root := range roots {
		if root.Path != currentPath {
			currentPath = root.Path
			fmt.Printf("\n%s\n", cmdutil.BoldWhite(currentPath))
		}
		printComment(root, 0)
		printReplies(threads, root.Comment.ID, 1)
	}

	// Print orphan replies (comments whose parent wasn't in our filtered set)
	for _, e := range entries {
		if e.Comment.InReplyTo != "" {
			found := false
			for _, r := range roots {
				if r.Comment.ID == e.Comment.InReplyTo {
					found = true
					break
				}
			}
			if !found {
				if _, printed := threads[e.Comment.ID]; !printed {
					if e.Path != currentPath {
						currentPath = e.Path
						fmt.Printf("\n%s\n", cmdutil.BoldWhite(currentPath))
					}
					printComment(e, 0)
				}
			}
		}
	}
}

func printComment(e commentEntry, indent int) {
	prefix := strings.Repeat("  ", indent)
	unresolvedTag := ""
	if e.Comment.Unresolved {
		unresolvedTag = cmdutil.Red(" [UNRESOLVED]")
	}
	line := ""
	if e.Comment.Line > 0 {
		line = fmt.Sprintf(":%d", e.Comment.Line)
	}
	fmt.Printf("%s  %s%s %s %s%s\n", prefix, cmdutil.Cyan(e.Comment.Author.DisplayName()),
		line, cmdutil.Dim("|"), cmdutil.FormatTimeAgo(e.Comment.Updated), unresolvedTag)
	for _, msgLine := range strings.Split(e.Comment.Message, "\n") {
		fmt.Printf("%s    %s\n", prefix, msgLine)
	}
}

func printReplies(threads map[string][]commentEntry, parentID string, indent int) {
	for _, reply := range threads[parentID] {
		printComment(reply, indent)
		printReplies(threads, reply.Comment.ID, indent+1)
	}
}
