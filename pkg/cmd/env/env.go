package env

import (
	"github.com/MakeNowJust/heredoc"
	cmdCreate "github.com/cli/cli/v2/pkg/cmd/env/create"
	cmdDelete "github.com/cli/cli/v2/pkg/cmd/env/delete"
	cmdExport "github.com/cli/cli/v2/pkg/cmd/env/export"
	cmdList "github.com/cli/cli/v2/pkg/cmd/env/list"
	cmdRun "github.com/cli/cli/v2/pkg/cmd/env/run"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdEnv(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env <command>",
		Short: "Manage GitHub Environments",
		Long: heredoc.Doc(`
			Env lists
`),
	}

	cmdutil.EnableRepoOverride(cmd, f)

	cmd.AddCommand(cmdExport.NewCmdExport(f, nil))
	cmd.AddCommand(cmdRun.NewCmdRun(f, nil))
	cmd.AddCommand(cmdCreate.NewCmdCreate(f, nil))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f, nil))
	cmd.AddCommand(cmdList.NewCmdList(f, nil))

	return cmd
}
