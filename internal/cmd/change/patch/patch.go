package patch

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
)

func NewCmdPatch() *cobra.Command {
	var (
		patchset int
		output   string
	)
	cmd := &cobra.Command{
		Use:   "patch <change-id>",
		Short: "Download patch file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			revision := "current"
			if patchset > 0 {
				revision = fmt.Sprintf("%d", patchset)
			}
			data, err := client.GetPatch(context.Background(), args[0], revision)
			cmdutil.ExitIfError(err)
			if output != "" {
				err = os.WriteFile(output, data, 0o644)
				cmdutil.ExitIfError(err)
				fmt.Printf("Patch saved to %s\n", output)
			} else {
				os.Stdout.Write(data)
			}
		},
	}
	cmd.Flags().IntVarP(&patchset, "patchset", "p", 0, "Patchset number (default: current)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (default: stdout)")
	return cmd
}
