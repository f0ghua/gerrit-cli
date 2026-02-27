package view

import (
	"fmt"
	"strings"

	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

func RenderDiffStat(files map[string]gerrit.FileInfo) {
	for name, f := range files {
		if name == "/COMMIT_MSG" {
			continue
		}
		total := f.LinesInserted + f.LinesDeleted
		bar := ""
		maxBar := 40
		if total > 0 {
			addLen := f.LinesInserted * maxBar / total
			delLen := maxBar - addLen
			if addLen > maxBar {
				addLen = maxBar
			}
			if delLen > maxBar {
				delLen = maxBar
			}
			bar = cmdutil.Green(strings.Repeat("+", addLen)) + cmdutil.Red(strings.Repeat("-", delLen))
		}
		fmt.Printf(" %-50s | %4d %s\n", cmdutil.TruncateString(name, 50), total, bar)
	}
}

func RenderDiff(diff *gerrit.DiffInfo, file string, noColor bool, contextLines int) {
	oldName := file
	newName := file
	if diff.MetaA.Name != "" {
		oldName = diff.MetaA.Name
	}
	if diff.MetaB.Name != "" {
		newName = diff.MetaB.Name
	}

	if noColor {
		fmt.Printf("--- a/%s\n+++ b/%s\n", oldName, newName)
	} else {
		fmt.Printf("%s\n%s\n", cmdutil.BoldWhite("--- a/"+oldName), cmdutil.BoldWhite("+++ b/"+newName))
	}

	lineA := 1
	lineB := 1

	for _, section := range diff.Content {
		if section.Skip > 0 {
			lineA += section.Skip
			lineB += section.Skip
			if noColor {
				fmt.Printf("... %d lines skipped ...\n", section.Skip)
			} else {
				fmt.Printf("%s\n", cmdutil.Dim(fmt.Sprintf("... %d lines skipped ...", section.Skip)))
			}
			continue
		}

		if len(section.A) > 0 || len(section.B) > 0 {
			countA := len(section.A)
			countB := len(section.B)
			hunk := fmt.Sprintf("@@ -%d,%d +%d,%d @@", lineA, countA, lineB, countB)
			if noColor {
				fmt.Println(hunk)
			} else {
				fmt.Println(cmdutil.Cyan(hunk))
			}
		}

		if len(section.AB) > 0 {
			lines := section.AB
			if contextLines >= 0 && len(section.A) == 0 && len(section.B) == 0 {
				if len(lines) > contextLines*2 {
					for i := 0; i < contextLines && i < len(lines); i++ {
						printContextLine(lines[i], noColor)
					}
					skip := len(lines) - contextLines*2
					if skip > 0 {
						if noColor {
							fmt.Printf("... %d lines skipped ...\n", skip)
						} else {
							fmt.Printf("%s\n", cmdutil.Dim(fmt.Sprintf("... %d lines skipped ...", skip)))
						}
					}
					for i := len(lines) - contextLines; i < len(lines); i++ {
						if i >= 0 {
							printContextLine(lines[i], noColor)
						}
					}
				} else {
					for _, l := range lines {
						printContextLine(l, noColor)
					}
				}
			} else {
				for _, l := range lines {
					printContextLine(l, noColor)
				}
			}
			lineA += len(lines)
			lineB += len(lines)
		}

		for _, l := range section.A {
			if noColor {
				fmt.Printf("- %s\n", l)
			} else {
				fmt.Println(cmdutil.Red("- " + l))
			}
			lineA++
		}

		for _, l := range section.B {
			if noColor {
				fmt.Printf("+ %s\n", l)
			} else {
				fmt.Println(cmdutil.Green("+ " + l))
			}
			lineB++
		}
	}
}

func printContextLine(line string, noColor bool) {
	if noColor {
		fmt.Printf("  %s\n", line)
	} else {
		fmt.Printf("  %s\n", cmdutil.Dim(line))
	}
}
