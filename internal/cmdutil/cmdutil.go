package cmdutil

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

var (
	BoldWhite  = color.New(color.FgWhite, color.Bold).SprintFunc()
	BoldGreen  = color.New(color.FgGreen, color.Bold).SprintFunc()
	BoldRed    = color.New(color.FgRed, color.Bold).SprintFunc()
	BoldYellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	BoldCyan   = color.New(color.FgCyan, color.Bold).SprintFunc()
	Green      = color.New(color.FgGreen).SprintFunc()
	Red        = color.New(color.FgRed).SprintFunc()
	Yellow     = color.New(color.FgYellow).SprintFunc()
	Cyan       = color.New(color.FgCyan).SprintFunc()
	Gray       = color.New(color.FgHiBlack).SprintFunc()
	Dim        = color.New(color.Faint).SprintFunc()
)

func ExitIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func FormatTimeAgo(timestamp string) string {
	formats := []string{
		"2006-01-02 15:04:05.000000000",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	var t time.Time
	for _, f := range formats {
		if parsed, err := time.Parse(f, timestamp); err == nil {
			t = parsed
			break
		}
	}
	if t.IsZero() {
		return Gray("unknown")
	}
	return Dim(timeAgo(t))
}

func timeAgo(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case diff < 24*time.Hour:
		h := int(diff.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case diff < 7*24*time.Hour:
		d := int(diff.Hours() / 24)
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	case diff < 30*24*time.Hour:
		w := int(diff.Hours() / (24 * 7))
		if w == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", w)
	case diff < 365*24*time.Hour:
		mo := int(diff.Hours() / (24 * 30))
		if mo == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", mo)
	default:
		y := int(diff.Hours() / (24 * 365))
		if y == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", y)
	}
}

func FormatScore(value int) string {
	switch {
	case value > 0:
		return BoldGreen(fmt.Sprintf("+%d", value))
	case value < 0:
		return BoldRed(fmt.Sprintf("%d", value))
	default:
		return Gray("0")
	}
}

func FormatChangeStatus(status string) string {
	switch strings.ToUpper(status) {
	case "NEW", "OPEN":
		return Green(status)
	case "MERGED":
		return BoldGreen(status)
	case "ABANDONED":
		return Red(status)
	default:
		return status
	}
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func PadString(s string, width int) string {
	visualLen := utf8.RuneCountInString(StripANSI(s))
	if visualLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visualLen)
}

func FormatTable(headers []string, rows [][]string, padding int) string {
	if len(rows) == 0 {
		return ""
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(StripANSI(h))
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				l := utf8.RuneCountInString(StripANSI(cell))
				if l > widths[i] {
					widths[i] = l
				}
			}
		}
	}
	var b strings.Builder
	for i, h := range headers {
		if i > 0 {
			b.WriteString(strings.Repeat(" ", padding))
		}
		b.WriteString(BoldWhite(PadString(h, widths[i])))
	}
	b.WriteString("\n")
	for i := range headers {
		if i > 0 {
			b.WriteString(strings.Repeat(" ", padding))
		}
		b.WriteString(strings.Repeat("-", widths[i]))
	}
	b.WriteString("\n")
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				b.WriteString(strings.Repeat(" ", padding))
			}
			if i < len(widths) {
				vl := utf8.RuneCountInString(StripANSI(cell))
				pad := widths[i] - vl
				b.WriteString(cell)
				if pad > 0 {
					b.WriteString(strings.Repeat(" ", pad))
				}
			} else {
				b.WriteString(cell)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func PatchsetRevision(change *gerrit.ChangeInfo, patchset int) string {
	if patchset <= 0 {
		return "current"
	}
	for _, rev := range change.Revisions {
		if rev.Number == patchset {
			return fmt.Sprintf("%d", patchset)
		}
	}
	return "current"
}
