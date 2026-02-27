package abandon

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

func NewCmdAbandon() *cobra.Command {
	var message string
	cmd := &cobra.Command{
		Use:   "abandon <change-id>",
		Short: "Abandon a change",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			change, err := client.Abandon(context.Background(), args[0], &gerrit.AbandonInput{Message: message})
			cmdutil.ExitIfError(err)
			fmt.Printf("Change %d abandoned.\n", change.Number)
		},
	}
	cmd.Flags().StringVarP(&message, "message", "m", "", "Abandon message")
	return cmd
}
