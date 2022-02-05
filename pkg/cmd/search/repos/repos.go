package repos

import (
	"fmt"
	"strings"

	"github.com/cli/cli/v2/pkg/cmd/search/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/export"
	"github.com/cli/cli/v2/utils"
	"github.com/spf13/cobra"
)

func NewCmdRepos(f *cmdutil.Factory) *cobra.Command {
	var webMode bool
	var template string
	var jq string
	query := shared.NewSearchQuery("repositories")

	cmd := &cobra.Command{
		Use:   "repos <query>",
		Short: "",
		Long:  "",
		RunE: func(c *cobra.Command, args []string) error {
			io := f.IOStreams
			query.Keywords = args
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			host, err := cfg.DefaultHost()
			if err != nil {
				return err
			}
			client, err := f.HttpClient()
			if err != nil {
				return err
			}
			searcher := shared.NewSearcher(host, client)
			if webMode {
				url := searcher.URL(query)
				if io.IsStdoutTTY() {
					fmt.Fprintf(io.ErrOut, "Opening %s in your browser.\n", utils.DisplayURL(url))
				}
				return f.Browser.Browse(url)
			}
			res, err := searcher.Search(query)
			if err != nil {
				return err
			}
			if jq != "" {
				err = export.FilterJSON(io.Out, strings.NewReader(res), jq)
				if err != nil {
					return err
				}
			} else if template != "" {
				t := export.NewTemplate(io, template)
				err = t.Execute(strings.NewReader(res))
				if err != nil {
					return err
				}
			} else {
				fmt.Println(res)
			}
			return nil
		},
	}

	//TODO: cant use jq and template at same time
	//TODO: add color json option?
	cmd.Flags().StringVarP(&jq, "jq", "q", "", "Query to select values from the response using jq syntax")
	cmd.Flags().StringVarP(&template, "template", "t", "", "Format the response using a Go template")
	cmd.Flags().BoolVarP(&webMode, "web", "w", false, "Open the search query in the web browser")

	cmd.Flags().BoolVar(&query.Paginate, "pageinate", false, "paginate")
	cmd.Flags().BoolVar(&query.Raw, "raw", false, "raw skip all qualifier flags")

	cmd.Flags().Var(query.Parameters.Order, "order", "order")
	cmd.Flags().Var(query.Parameters.Page, "page", "page")
	cmd.Flags().Var(query.Parameters.PerPage, "per-page", "per-page")
	cmd.Flags().Var(query.Parameters.Sort, "sort", "sort")

	cmd.Flags().Var(query.Qualifiers["Archived"], "archived", "archived")
	cmd.Flags().Var(query.Qualifiers["Created"], "created", "created at")
	cmd.Flags().Var(query.Qualifiers["Followers"], "followers", "followers")
	cmd.Flags().Var(query.Qualifiers["Fork"], "fork", "forks")
	cmd.Flags().Var(query.Qualifiers["Forks"], "forks", "forks")
	cmd.Flags().Var(query.Qualifiers["GoodFirstIssues"], "good-first-issues", "good-first-issues")
	cmd.Flags().Var(query.Qualifiers["HelpWantedIssues"], "help-wanted-issues", "help-wanted-issues")
	cmd.Flags().Var(query.Qualifiers["In"], "in", "in")
	cmd.Flags().Var(query.Qualifiers["Language"], "language", "language")
	cmd.Flags().Var(query.Qualifiers["License"], "license", "license")
	cmd.Flags().Var(query.Qualifiers["Mirror"], "mirror", "mirror")
	cmd.Flags().Var(query.Qualifiers["Org"], "org", "org")
	cmd.Flags().Var(query.Qualifiers["Pushed"], "pushed", "pushed at")
	cmd.Flags().Var(query.Qualifiers["Repo"], "repo", "repo")
	cmd.Flags().Var(query.Qualifiers["Size"], "size", "size")
	cmd.Flags().Var(query.Qualifiers["Stars"], "stars", "stars")
	cmd.Flags().Var(query.Qualifiers["Topic"], "topic", "topic")
	cmd.Flags().Var(query.Qualifiers["Topics"], "topics", "number topics")
	cmd.Flags().Var(query.Qualifiers["User"], "user", "user")
	cmd.Flags().Var(query.Qualifiers["Visibility"], "is", "is")

	return cmd
}
