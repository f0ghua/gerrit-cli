package view

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	viewPkg "github.com/fog/gerrit-cli/internal/view"
)

func NewCmdView() *cobra.Command {
	var (
		showFiles     bool
		showPatchsets bool
		jsonOut       bool
	)
	cmd := &cobra.Command{
		Use:   "view <change-id>",
		Short: "View change details",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			opts := []string{"LABELS", "CURRENT_REVISION", "CURRENT_COMMIT", "DETAILED_ACCOUNTS", "DETAILED_LABELS"}
			if showPatchsets {
				opts = []string{"LABELS", "ALL_REVISIONS", "ALL_COMMITS", "DETAILED_ACCOUNTS", "DETAILED_LABELS"}
			}
			change, err := client.GetChange(context.Background(), args[0], opts...)
			cmdutil.ExitIfError(err)
			if jsonOut {
				data, _ := json.MarshalIndent(change, "", "  ")
				fmt.Println(string(data))
				return
			}
			viewPkg.RenderChangeDetail(change, showFiles, showPatchsets)
			if showFiles {
				files, err := client.GetChangeFiles(context.Background(), args[0], "current")
				cmdutil.ExitIfError(err)
				viewPkg.RenderFiles(files)
			}
		},
	}
	cmd.Flags().BoolVar(&showFiles, "files", false, "Show changed files")
	cmd.Flags().BoolVar(&showPatchsets, "patchsets", false, "Show patchset history")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON output")
	return cmd
}
