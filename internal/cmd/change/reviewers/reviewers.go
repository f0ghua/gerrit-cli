package reviewers

import (
	"github.com/spf13/cobra"

	"github.com/fog/gerrit-cli/internal/cmd/change/reviewers/add"
	"github.com/fog/gerrit-cli/internal/cmd/change/reviewers/remove"
)

func NewCmdReviewers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reviewers",
		Short: "Manage reviewers on a change",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(add.NewCmdAdd(), remove.NewCmdRemove())
	return cmd
}
