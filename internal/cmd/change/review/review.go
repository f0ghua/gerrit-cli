package review

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

func NewCmdReview() *cobra.Command {
	var (
		patchset   int
		message    string
		codeReview int
		verified   int
		labels     []string
		comments   []string
	)
	cmd := &cobra.Command{
		Use:   "review <change-id>",
		Short: "Post a review on a change",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			revision := "current"
			if patchset > 0 {
				revision = fmt.Sprintf("%d", patchset)
			}
			input := &gerrit.ReviewInput{
				Message: message,
				Labels:  make(map[string]int),
			}
			if cmd.Flags().Changed("code-review") {
				input.Labels["Code-Review"] = codeReview
			}
			if cmd.Flags().Changed("verified") {
				input.Labels["Verified"] = verified
			}
			for _, l := range labels {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					score, err := strconv.Atoi(parts[1])
					if err == nil {
						input.Labels[parts[0]] = score
					}
				}
			}
			if len(comments) > 0 {
				cm, err := parseComments(comments)
				cmdutil.ExitIfError(err)
				input.Comments = cm
			}
			err := client.SetReview(context.Background(), args[0], revision, input)
			cmdutil.ExitIfError(err)
			fmt.Println("Review posted.")
		},
	}
	cmd.Flags().IntVarP(&patchset, "patchset", "p", 0, "Patchset number (default: current)")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Review message")
	cmd.Flags().IntVar(&codeReview, "code-review", 0, "Code-Review score (-2 to +2)")
	cmd.Flags().IntVar(&verified, "verified", 0, "Verified score (-1 to +1)")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Label score (Name=score)")
	cmd.Flags().StringArrayVarP(&comments, "comment", "c", nil, "Inline comment (file:line:message)")
	return cmd
}

func parseComments(raw []string) (map[string][]gerrit.CommentInput, error) {
	result := make(map[string][]gerrit.CommentInput)
	for _, s := range raw {
		parts := strings.SplitN(s, ":", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid comment format %q: expected file:line:message", s)
		}
		line, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid line number in comment %q: %w", s, err)
		}
		result[parts[0]] = append(result[parts[0]], gerrit.CommentInput{
			Line:    line,
			Message: parts[2],
		})
	}
	return result, nil
}
