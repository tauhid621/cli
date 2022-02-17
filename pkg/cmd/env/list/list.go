package list

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/ghinstance"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	Config     func() (config.Config, error)
	BaseRepo   func() (ghrepo.Interface, error)
}

func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "list <EnvName>",
		Short: "lists env",
		Long: heredoc.Doc(`
			Lists env in a json
		`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// support `-R, --repo` override
			opts.BaseRepo = f.BaseRepo

			if runF != nil {
				return runF(opts)
			}

			return listRun(opts)
		},
	}

	return cmd
}

func listRun(opts *ListOptions) error {
	client, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}

	var baseRepo ghrepo.Interface
	baseRepo, err = opts.BaseRepo()

	envs, err := GetEnvs(client, baseRepo)
	if err != nil {
		return fmt.Errorf("failed to get env: %w", err)
	}

	// fmt.Println(string(envs[0].Name))

	jsonContent, _ := json.MarshalIndent(envs, "", " ")
	fmt.Println(string(jsonContent))

	return nil
}

type Env struct {
	Id        int
	Name      string
	Variables map[string]string
	Url       string
	UpdatedAt time.Time `json:"updated_at"`
}

type envPayload struct {
	TotalCount   uint16 `json:"total_count"`
	Environments []Env
}

func GetEnvs(client httpClient, repo ghrepo.Interface) ([]Env, error) {

	path := fmt.Sprintf("repos/%s/environments", ghrepo.FullName(repo))
	url := fmt.Sprintf("%s%s", ghinstance.RESTPrefix(repo.RepoHost()), path)
	//
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, api.HandleHTTPError(resp)
	}

	// b, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, err
	// }

	// var r envPayload
	// err = json.Unmarshal(b, &r)
	// if err != nil {
	// 	return nil, err
	// }

	// fmt.Println(b)
	// fmt.Println(r)

	dec := json.NewDecoder(resp.Body)

	var envs envPayload

	if err := dec.Decode(&envs); err != nil {
		return nil, err
	}
	return envs.Environments, nil
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}
