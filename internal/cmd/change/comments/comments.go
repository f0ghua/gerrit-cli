package comments

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/internal/view"
)

func NewCmdComments() *cobra.Command {
	var (
		patchset int
		all      bool
		jsonOut  bool
	)
	cmd := &cobra.Command{
		Use:   "comments <change-id>",
		Short: "View comments on a change",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			comments, err := client.GetComments(context.Background(), args[0])
			cmdutil.ExitIfError(err)
			if jsonOut {
				data, _ := json.MarshalIndent(comments, "", "  ")
				fmt.Println(string(data))
				return
			}
			view.RenderComments(comments, all, patchset)
		},
	}
	cmd.Flags().IntVarP(&patchset, "patchset", "p", 0, "Filter by patchset number")
	cmd.Flags().BoolVar(&all, "all", false, "Show all comments (default: unresolved only)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON output")
	return cmd
}
