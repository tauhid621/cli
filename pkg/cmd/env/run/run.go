package env

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"github.com/MakeNowJust/heredoc"
	cmdExport "github.com/cli/cli/v2/pkg/cmd/env/export"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type RunOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	Config     func() (config.Config, error)
	BaseRepo   func() (ghrepo.Interface, error)

	EnvName string
	Command []string
}

func NewCmdRun(f *cmdutil.Factory, runF func(*RunOptions) error) *cobra.Command {
	opts := &RunOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "run <EnvName> <command>",
		Short: "runs command in the env",
		Long: heredoc.Doc(`
			Runs the command in env 
		`),
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// support `-R, --repo` override
			opts.BaseRepo = f.BaseRepo
			opts.EnvName = args[0]
			opts.Command = args[1:]

			if runF != nil {
				return runF(opts)
			}

			return runRun(opts)
		},
	}

	return cmd
}

func runRun(opts *RunOptions) error {
	client, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}

	envName := opts.EnvName

	var baseRepo ghrepo.Interface
	baseRepo, err = opts.BaseRepo()

	var secrets []*cmdExport.Secret
	secrets, err = cmdExport.GetEnvSecrets(client, baseRepo, envName)
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	env, err := cmdExport.GetEnv(client, baseRepo, envName)
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	envVar := os.Environ()

	for _, secret := range secrets {
		envVar = append(envVar, fmt.Sprintf("%s=%s", secret.Name, secret.Value))
	}

	for _, variable := range env.Variables {
		envVar = append(envVar, fmt.Sprintf("%s=%s", variable.Name, variable.Value))
	}

	binary, lookErr := exec.LookPath(opts.Command[0])
	if lookErr != nil {
		panic(lookErr)
	}

	// args := []string{opts.Command}

	execErr := syscall.Exec(binary, opts.Command, envVar)
	if execErr != nil {
		panic(execErr)
	}

	fmt.Println("main code : should not be reachable ")

	return nil
}
