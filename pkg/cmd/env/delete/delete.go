package delete

import (
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/ghinstance"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	Config     func() (config.Config, error)
	BaseRepo   func() (ghrepo.Interface, error)

	EnvName string
}

// type GhEnvironment struct {
// 	Name      string
// 	Variables string
// }

func NewCmdDelete(f *cmdutil.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "delete <EnvName>",
		Short: "deletes env",
		Long: heredoc.Doc(`
			Deletes env in a json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// support `-R, --repo` override
			opts.BaseRepo = f.BaseRepo
			opts.EnvName = args[0]

			if runF != nil {
				return runF(opts)
			}

			return deleteRun(opts)
		},
	}

	return cmd
}

func deleteRun(opts *DeleteOptions) error {
	client, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("could not delete http client: %w", err)
	}

	envName := opts.EnvName

	var baseRepo ghrepo.Interface
	baseRepo, err = opts.BaseRepo()

	err = DeleteOrUpdateEnv(client, baseRepo, envName, opts)
	if err != nil {
		return fmt.Errorf("error while creating/updating env: %w", err)
	}

	return nil
}

func DeleteOrUpdateEnv(client httpClient, repo ghrepo.Interface, envName string, deleteOptions *DeleteOptions) error {
	path := fmt.Sprintf("repos/%s/environments/%s", ghrepo.FullName(repo), envName)

	req, err := http.NewRequest("DELETE", ghinstance.RESTPrefix(repo.RepoHost())+path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return api.HandleHTTPError(resp)
	}

	return nil
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}
