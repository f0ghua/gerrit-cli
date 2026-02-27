package add

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

func NewCmdAdd() *cobra.Command {
	var cc bool
	cmd := &cobra.Command{
		Use:   "add <change-id> <account>",
		Short: "Add a reviewer to a change",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			state := "REVIEWER"
			if cc {
				state = "CC"
			}
			result, err := client.AddReviewer(context.Background(), args[0], &gerrit.ReviewerInput{
				Reviewer: args[1],
				State:    state,
			})
			cmdutil.ExitIfError(err)
			if result.Error != "" {
				cmdutil.ExitIfError(fmt.Errorf("%s", result.Error))
			}
			fmt.Printf("Added %s as %s.\n", args[1], state)
		},
	}
	cmd.Flags().BoolVar(&cc, "cc", false, "Add as CC instead of reviewer")
	return cmd
}
