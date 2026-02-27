package view

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

func RenderChangeList(changes []gerrit.ChangeInfo, plain, noHeaders bool) {
	if len(changes) == 0 {
		fmt.Println("No changes found.")
		return
	}
	headers := []string{"ID", "Subject", "Owner", "Project", "Branch", "Status", "Updated", "+/-"}
	var rows [][]string
	for _, c := range changes {
		rows = append(rows, []string{
			fmt.Sprintf("%d", c.Number),
			cmdutil.TruncateString(c.Subject, 50),
			c.Owner.DisplayName(),
			c.Project,
			c.Branch,
			cmdutil.FormatChangeStatus(c.Status),
			cmdutil.FormatTimeAgo(c.Updated),
			fmt.Sprintf("+%d/-%d", c.Insertions, c.Deletions),
		})
	}
	if plain {
		if !noHeaders {
			fmt.Println(strings.Join(headers, "\t"))
		}
		for _, row := range rows {
			fmt.Println(strings.Join(row, "\t"))
		}
		return
	}
	if noHeaders {
		for _, row := range rows {
			fmt.Println(strings.Join(row, "  "))
		}
		return
	}
	fmt.Print(cmdutil.FormatTable(headers, rows, 2))
}

func RenderChangeDetail(change *gerrit.ChangeInfo, showFiles bool, showPatchsets bool) {
	fmt.Printf("%s %s\n", cmdutil.BoldWhite(fmt.Sprintf("#%d", change.Number)), cmdutil.BoldWhite(change.Subject))
	fmt.Printf("Status:   %s\n", cmdutil.FormatChangeStatus(change.Status))
	fmt.Printf("Owner:    %s\n", change.Owner.DisplayName())
	fmt.Printf("Project:  %s (%s)\n", change.Project, change.Branch)
	fmt.Printf("Created:  %s\n", cmdutil.FormatTimeAgo(change.Created))
	fmt.Printf("Updated:  %s\n", cmdutil.FormatTimeAgo(change.Updated))
	fmt.Printf("Changes:  +%d/-%d\n", change.Insertions, change.Deletions)

	if len(change.Labels) > 0 {
		fmt.Println("\nLabels:")
		for name, label := range change.Labels {
			var votes []string
			for _, a := range label.All {
				if a.Value != 0 {
					votes = append(votes, fmt.Sprintf("%s %s", a.DisplayName(), cmdutil.FormatScore(a.Value)))
				}
			}
			if len(votes) > 0 {
				fmt.Printf("  %-20s %s\n", name+":", strings.Join(votes, ", "))
			} else {
				fmt.Printf("  %-20s %s\n", name+":", cmdutil.Gray("no votes"))
			}
		}
	}

	if reviewers, ok := change.Reviewers["REVIEWER"]; ok && len(reviewers) > 0 {
		fmt.Println("\nReviewers:")
		for _, r := range reviewers {
			fmt.Printf("  %s\n", r.DisplayName())
		}
	}

	if showPatchsets && len(change.Revisions) > 0 {
		fmt.Println("\nPatchsets:")
		type psEntry struct {
			Number int
			SHA    string
			Rev    gerrit.RevisionInfo
		}
		var entries []psEntry
		for sha, rev := range change.Revisions {
			entries = append(entries, psEntry{rev.Number, sha, rev})
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Number < entries[j].Number })
		headers := []string{"PS", "SHA", "Author", "Created", "Subject"}
		var rows [][]string
		for _, e := range entries {
			sha := e.SHA
			if len(sha) > 8 {
				sha = sha[:8]
			}
			rows = append(rows, []string{
				fmt.Sprintf("%d", e.Number),
				sha,
				e.Rev.Commit.Author.Name,
				cmdutil.FormatTimeAgo(e.Rev.Created),
				cmdutil.TruncateString(e.Rev.Commit.Subject, 50),
			})
		}
		fmt.Print(cmdutil.FormatTable(headers, rows, 2))
	}
}

func RenderFiles(files map[string]gerrit.FileInfo) {
	if len(files) == 0 {
		return
	}
	fmt.Println("\nFiles:")
	for name, f := range files {
		if name == "/COMMIT_MSG" {
			continue
		}
		status := " "
		switch f.Status {
		case "A":
			status = cmdutil.Green("A")
		case "D":
			status = cmdutil.Red("D")
		case "R":
			status = cmdutil.Yellow("R")
		default:
			status = cmdutil.Cyan("M")
		}
		fmt.Printf("  %s %s (+%d/-%d)\n", status, name, f.LinesInserted, f.LinesDeleted)
	}
}
