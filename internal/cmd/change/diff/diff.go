package diff

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/internal/view"
)

func NewCmdDiff() *cobra.Command {
	var (
		patchset     int
		base         int
		file         string
		contextLines int
		noColor      bool
		stat         bool
	)
	cmd := &cobra.Command{
		Use:   "diff <change-id>",
		Short: "View file diffs for a change",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			revision := "current"
			if patchset > 0 {
				revision = fmt.Sprintf("%d", patchset)
			}
			files, err := client.GetChangeFiles(context.Background(), args[0], revision)
			cmdutil.ExitIfError(err)

			if stat {
				view.RenderDiffStat(files)
				return
			}

			if file != "" {
				d, err := client.GetFileDiff(context.Background(), args[0], revision, file, base)
				cmdutil.ExitIfError(err)
				view.RenderDiff(d, file, noColor, contextLines)
				return
			}

			for name := range files {
				if name == "/COMMIT_MSG" {
					continue
				}
				d, err := client.GetFileDiff(context.Background(), args[0], revision, name, base)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error fetching diff for %s: %s\n", name, err)
					continue
				}
				view.RenderDiff(d, name, noColor, contextLines)
				fmt.Println()
			}
		},
	}
	cmd.Flags().IntVarP(&patchset, "patchset", "p", 0, "Patchset number (default: current)")
	cmd.Flags().IntVar(&base, "base", 0, "Base patchset for inter-patchset diff")
	cmd.Flags().StringVarP(&file, "file", "f", "", "Show diff for specific file")
	cmd.Flags().IntVarP(&contextLines, "context", "C", 3, "Context lines around changes")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable color output")
	cmd.Flags().BoolVar(&stat, "stat", false, "Show diffstat summary only")
	return cmd
}
