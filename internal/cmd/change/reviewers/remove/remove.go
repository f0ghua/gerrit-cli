package remove

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
)

func NewCmdRemove() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <change-id> <account>",
		Short: "Remove a reviewer from a change",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			err := client.RemoveReviewer(context.Background(), args[0], args[1])
			cmdutil.ExitIfError(err)
			fmt.Printf("Removed %s from reviewers.\n", args[1])
		},
	}
}
