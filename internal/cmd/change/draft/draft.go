package draft

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

func NewCmdDraft() *cobra.Command {
	var (
		patchset int
		comments []string
	)
	cmd := &cobra.Command{
		Use:   "draft <change-id>",
		Short: "Save inline comments as drafts",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			revision := "current"
			if patchset > 0 {
				revision = fmt.Sprintf("%d", patchset)
			}
			for _, s := range comments {
				parts := strings.SplitN(s, ":", 3)
				if len(parts) != 3 {
					cmdutil.ExitIfError(fmt.Errorf("invalid comment format %q: expected file:line:message", s))
				}
				line, err := strconv.Atoi(parts[1])
				if err != nil {
					cmdutil.ExitIfError(fmt.Errorf("invalid line number in comment %q: %w", s, err))
				}
				input := &gerrit.DraftInput{
					Path:    parts[0],
					Line:    line,
					Message: parts[2],
				}
				_, err = client.CreateDraft(context.Background(), args[0], revision, input)
				cmdutil.ExitIfError(err)
			}
			fmt.Printf("Saved %d draft comment(s).\n", len(comments))
		},
	}
	cmd.Flags().IntVarP(&patchset, "patchset", "p", 0, "Patchset number (default: current)")
	cmd.Flags().StringArrayVarP(&comments, "comment", "c", nil, "Inline comment (file:line:message)")
	_ = cmd.MarkFlagRequired("comment")
	return cmd
}
