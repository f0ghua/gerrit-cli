package drafts

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
)

func NewCmdDrafts() *cobra.Command {
	var patchset int
	cmd := &cobra.Command{
		Use:   "drafts <change-id>",
		Short: "List draft comments",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			revision := "current"
			if patchset > 0 {
				revision = fmt.Sprintf("%d", patchset)
			}
			drafts, err := client.GetDrafts(context.Background(), args[0], revision)
			cmdutil.ExitIfError(err)
			if len(drafts) == 0 {
				fmt.Println("No drafts.")
				return
			}
			files := make([]string, 0, len(drafts))
			for f := range drafts {
				files = append(files, f)
			}
			sort.Strings(files)
			for _, f := range files {
				for _, d := range drafts[f] {
					fmt.Printf("%s:%d — %s\n", f, d.Line, d.Message)
				}
			}
		},
	}
	cmd.Flags().IntVarP(&patchset, "patchset", "p", 0, "Patchset number (default: current)")
	return cmd
}
