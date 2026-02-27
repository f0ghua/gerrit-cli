package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/api"
	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/internal/view"
	"github.com/fog/gerrit-cli/pkg/query"
)

func NewCmdList() *cobra.Command {
	var (
		status    string
		reviewer  bool
		project   string
		branch    string
		rawQuery  string
		limit     int
		plain     bool
		noHeaders bool
		jsonOut   bool
	)
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List changes",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			client := api.DefaultClient(viper.GetBool("debug"))
			q := query.New()
			if rawQuery != "" {
				q.Raw(rawQuery)
			} else {
				if status != "" {
					q.Status(status)
				}
				if reviewer {
					q.Reviewer("self")
				}
				if project == "" {
					project = viper.GetString("project")
				}
				if project != "" {
					q.Project(project)
				}
				if branch != "" {
					q.Branch(branch)
				}
			}
			changes, err := client.ListChanges(context.Background(), q.String(), limit)
			cmdutil.ExitIfError(err)
			if jsonOut {
				data, _ := json.MarshalIndent(changes, "", "  ")
				fmt.Println(string(data))
				return
			}
			view.RenderChangeList(changes, plain, noHeaders)
		},
	}
	cmd.Flags().StringVarP(&status, "status", "s", "open", "Filter by status")
	cmd.Flags().BoolVar(&reviewer, "reviewer", false, "Show changes where you are a reviewer")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Filter by project")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Filter by branch")
	cmd.Flags().StringVarP(&rawQuery, "query", "q", "", "Raw Gerrit query")
	cmd.Flags().IntVarP(&limit, "limit", "n", 25, "Max results")
	cmd.Flags().BoolVar(&plain, "plain", false, "Plain output (tab-separated)")
	cmd.Flags().BoolVar(&noHeaders, "no-headers", false, "Hide headers")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "JSON output")
	return cmd
}
