package submit

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
)

func NewCmdSubmit() *cobra.Command {
	return &cobra.Command{
		Use:   "submit <change-id>",
		Short: "Submit a change",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			change, err := client.Submit(context.Background(), args[0])
			cmdutil.ExitIfError(err)
			fmt.Printf("Change %d submitted (%s).\n", change.Number, change.Status)
		},
	}
}
