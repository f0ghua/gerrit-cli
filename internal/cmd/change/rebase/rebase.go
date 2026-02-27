package rebase

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

func NewCmdRebase() *cobra.Command {
	var (
		base           string
		allowConflicts bool
	)
	cmd := &cobra.Command{
		Use:   "rebase <change-id>",
		Short: "Rebase a change",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			change, err := client.Rebase(context.Background(), args[0], &gerrit.RebaseInput{
				Base:           base,
				AllowConflicts: allowConflicts,
			})
			cmdutil.ExitIfError(err)
			fmt.Printf("Change %d rebased.\n", change.Number)
		},
	}
	cmd.Flags().StringVarP(&base, "base", "b", "", "Base commit or change to rebase onto")
	cmd.Flags().BoolVar(&allowConflicts, "allow-conflicts", false, "Allow rebase with conflicts")
	return cmd
}
