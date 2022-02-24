package create

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/ghinstance"
	"github.com/cli/cli/v2/internal/ghrepo"
	cmdExport "github.com/cli/cli/v2/pkg/cmd/env/export"
	secretSet "github.com/cli/cli/v2/pkg/cmd/secret/set"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type CreateOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	Config     func() (config.Config, error)
	BaseRepo   func() (ghrepo.Interface, error)

	EnvName   string
	Variables string
	Secrets   string
	Who       string
}

// type GhEnvironment struct {
// 	Name      string
// 	Variables string
// }

func NewCmdCreate(f *cmdutil.Factory, runF func(*CreateOptions) error) *cobra.Command {
	opts := &CreateOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "create <EnvName>",
		Short: "creates env",
		Long: heredoc.Doc(`
			Creates env in a json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// support `-R, --repo` override
			opts.BaseRepo = f.BaseRepo
			opts.EnvName = args[0]

			if runF != nil {
				return runF(opts)
			}

			return createRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Variables, "variables", "v", "", "Variables for the env")
	cmd.Flags().StringVarP(&opts.Secrets, "secrets", "s", "", "Variables for the env")
	cmd.Flags().StringVarP(&opts.Who, "who", "w", "", "who can access the env")

	return cmd
}

func createRun(opts *CreateOptions) error {
	client, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}

	envName := opts.EnvName

	var baseRepo ghrepo.Interface
	baseRepo, err = opts.BaseRepo()

	err = CreateOrUpdateEnv(client, baseRepo, envName, opts)
	if err != nil {
		return fmt.Errorf("error while creating/updating env: %w", err)
	}

	sclient := api.NewClientFromHTTP(client)

	if opts.Secrets != "" {
		secrets_map, err := getStringMap(opts.Secrets)
		if err != nil {
			return err
		}

		for key, value := range secrets_map {
			// fmt.Println("Key:", key, "=>", "Element:", value)

			secretSet.SetEnvSecret(baseRepo.RepoHost(), sclient, baseRepo, envName, key, []byte(value))
		}

	}

	// export the env
	exportOpts := &cmdExport.ExportOptions{
		IO:         opts.IO,
		Config:     opts.Config,
		HttpClient: opts.HttpClient,
		BaseRepo:   opts.BaseRepo,
		EnvName:    opts.EnvName,
	}

	cmdExport.ExportRun(exportOpts)

	return nil
}

func CreateOrUpdateEnv(client httpClient, repo ghrepo.Interface, envName string, createOptions *CreateOptions) error {
	path := fmt.Sprintf("repos/%s/environments/%s", ghrepo.FullName(repo), envName)

	var reqBody []byte
	var err error
	if createOptions.Variables != "" {
		variables_map, err := getStringMap(createOptions.Variables)
		if err != nil {
			return err
		}

		reqBody, err = json.Marshal(map[string]interface{}{"variables": variables_map})
		if err != nil {
			return err
		}
	}

	if createOptions.Who != "" {
		// doing [1:] as expecting @ before name

		split := strings.Split(createOptions.Who, "@")
		access := split[0]
		user := split[1]

		if access == "read" {
			reqBody, err = json.Marshal(map[string]interface{}{"access": user})
		} else if access == "read_secrets" {
			reqBody, err = json.Marshal(map[string]interface{}{"secretsaccess": user})
		} else {
			fmt.Println("invalid access for user")
		}

		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest("PUT", ghinstance.RESTPrefix(repo.RepoHost())+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return api.HandleHTTPError(res)
	}

	if res.Body != nil {
		_, _ = io.Copy(ioutil.Discard, res.Body)
	}

	return nil
}

func getStringMap(variables string) (map[string]string, error) {
	vars := strings.Split(variables, ",")
	varMap := make(map[string]string)

	for _, envVar := range vars {
		s := strings.Split(envVar, "=")
		varMap[strings.Trim(s[0], " ")] = strings.Trim(s[1], " ")
	}

	return varMap, nil
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}
