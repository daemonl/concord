package cmd

import (
	"io"
	"os"

	"github.com/gomicro/concord/report"
	"github.com/spf13/cobra"
)

var applyCmd = NewApplyCmd(os.Stdout)

func init() {
	rootCmd.AddCommand(applyCmd)
}

func NewApplyCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "apply",
		Args:              cobra.ExactArgs(1),
		Short:             "Apply an org configuration",
		Long:              `Apply an org configuration against github`,
		PersistentPreRunE: setupClient,
		RunE:              applyRun,
	}

	cmd.SetOut(out)

	return cmd
}

func applyRun(cmd *cobra.Command, args []string) error {
	file := args[0]

	org, err := readManifest(file)
	if err != nil {
		return handleError(cmd, err)
	}

	report.PrintHeader("Org")
	report.Println()

	err = membersRun(cmd, args, org, false)
	if err != nil {
		return handleError(cmd, err)
	}

	err = teamsRun(cmd, args, org, false)
	if err != nil {
		return handleError(cmd, err)
	}

	err = reposRun(cmd, args, org, false)
	if err != nil {
		return handleError(cmd, err)
	}

	return nil
}
