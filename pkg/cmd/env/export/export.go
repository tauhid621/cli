package export

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/ghinstance"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/cli/v2/utils"
	"github.com/spf13/cobra"
)

type ExportOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	Config     func() (config.Config, error)
	BaseRepo   func() (ghrepo.Interface, error)

	EnvName string
	Format  string
}

type GhEnvironment struct {
	Name    string
	Secrets []*Secret
}

func NewCmdExport(f *cmdutil.Factory, runF func(*ExportOptions) error) *cobra.Command {
	opts := &ExportOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "export <EnvName>",
		Short: "exports env",
		Long: heredoc.Doc(`
			Exports env in a json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// support `-R, --repo` override
			opts.BaseRepo = f.BaseRepo
			opts.EnvName = args[0]

			if runF != nil {
				return runF(opts)
			}

			return exportRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", "", "List secrets for an organization")

	return cmd
}

func exportRun(opts *ExportOptions) error {
	client, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}

	envName := opts.EnvName

	var baseRepo ghrepo.Interface
	baseRepo, err = opts.BaseRepo()

	var secrets []*Secret

	secrets, err = GetEnvSecrets(client, baseRepo, envName)

	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	if opts.Format == "" {
		if err := opts.IO.StartPager(); err == nil {
			defer opts.IO.StopPager()
		} else {
			fmt.Fprintf(opts.IO.ErrOut, "failed to start pager: %v\n", err)
		}

		tp := utils.NewTablePrinter(opts.IO)
		for _, secret := range secrets {
			tp.AddField(secret.Name, nil, nil)
			tp.AddField(string(secret.Value), nil, nil)
			updatedAt := secret.UpdatedAt.Format("2006-01-02")
			if opts.IO.IsStdoutTTY() {
				updatedAt = fmt.Sprintf("Updated %s", updatedAt)
			}
			tp.AddField(updatedAt, nil, nil)

			tp.EndRow()
		}

		err = tp.Render()
		if err != nil {
			return err
		}
	} else if opts.Format == "json" {
		env := &GhEnvironment{Name: envName, Secrets: secrets}

		jsonContent, _ := json.MarshalIndent(env, "", " ")
		fmt.Println(string(jsonContent))
		// _ = ioutil.WriteFile(opts.JsonFileName, fileContent, 0644)

		// fmt.Printf("Environment exported to %s\n", opts.JsonFileName)
	} else {
		fmt.Println("Format not supported")
	}

	return nil
}

type Secret struct {
	Name      string
	Value     string
	UpdatedAt time.Time `json:"updated_at"`
}

func GetEnvSecrets(client httpClient, repo ghrepo.Interface, envName string) ([]*Secret, error) {
	path := fmt.Sprintf("repos/%s/environments/%s/secrets", ghrepo.FullName(repo), envName)
	return getSecrets(client, repo.RepoHost(), path)
}

type secretsPayload struct {
	Secrets []*Secret
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func getSecrets(client httpClient, host, path string) ([]*Secret, error) {
	var results []*Secret
	url := fmt.Sprintf("%s%s?per_page=100", ghinstance.RESTPrefix(host), path)

	for {
		var payload secretsPayload
		nextURL, err := apiGet(client, url, &payload)
		if err != nil {
			return nil, err
		}
		results = append(results, payload.Secrets...)

		if nextURL == "" {
			break
		}
		url = nextURL
	}

	return results, nil
}

func apiGet(client httpClient, url string, data interface{}) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return "", api.HandleHTTPError(resp)
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(data); err != nil {
		return "", err
	}

	return findNextPage(resp.Header.Get("Link")), nil
}

var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)

func findNextPage(link string) string {
	for _, m := range linkRE.FindAllStringSubmatch(link, -1) {
		if len(m) > 2 && m[2] == "next" {
			return m[1]
		}
	}
	return ""
}
