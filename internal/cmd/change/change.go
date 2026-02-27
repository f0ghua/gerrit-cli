package change

import (
	"github.com/spf13/cobra"

	"github.com/fog/gerrit-cli/internal/cmd/change/abandon"
	"github.com/fog/gerrit-cli/internal/cmd/change/comments"
	"github.com/fog/gerrit-cli/internal/cmd/change/diff"
	"github.com/fog/gerrit-cli/internal/cmd/change/list"
	"github.com/fog/gerrit-cli/internal/cmd/change/patch"
	"github.com/fog/gerrit-cli/internal/cmd/change/rebase"
	"github.com/fog/gerrit-cli/internal/cmd/change/review"
	"github.com/fog/gerrit-cli/internal/cmd/change/reviewers"
	"github.com/fog/gerrit-cli/internal/cmd/change/submit"
	"github.com/fog/gerrit-cli/internal/cmd/change/view"
)

func NewCmdChange() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "change",
		Short:   "Manage Gerrit changes",
		Long:    "Commands for listing, viewing, and managing Gerrit code review changes.",
		Aliases: []string{"changes"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		list.NewCmdList(),
		view.NewCmdView(),
		comments.NewCmdComments(),
		review.NewCmdReview(),
		submit.NewCmdSubmit(),
		abandon.NewCmdAbandon(),
		rebase.NewCmdRebase(),
		diff.NewCmdDiff(),
		patch.NewCmdPatch(),
		reviewers.NewCmdReviewers(),
	)
	return cmd
}
