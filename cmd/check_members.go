package cmd

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/gomicro/concord/client"
	gh_pb "github.com/gomicro/concord/github/v1"
	"github.com/gomicro/concord/report"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func init() {
	checkCmd.AddCommand(NewCheckMembersCmd(os.Stdout))
}

func NewCheckMembersCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Args:  cobra.ExactArgs(1),
		Short: "Check members exists in an organization",
		Long:  `Check members in a configuration against what exists in github`,
		RunE:  checkMembersRun,
	}

	cmd.SetOut(out)

	return cmd
}

func checkMembersRun(cmd *cobra.Command, args []string) error {
	file := args[0]

	org, err := readManifest(file)
	if err != nil {
		return handleError(cmd, err)
	}

	report.PrintHeader("Org")
	report.Println()

	return membersRun(cmd, args, org, true)
}

func membersRun(cmd *cobra.Command, args []string, org *gh_pb.Organization, dry bool) error {
	ctx := cmd.Context()

	clt, err := client.ClientFromContext(ctx)
	if err != nil {
		return handleError(cmd, err)
	}

	report.Println()
	report.PrintHeader("Members")
	report.Println()

	// check people exist
	ps, err := clt.GetMembers(ctx, org.Name)
	if err != nil {
		return handleError(cmd, err)
	}

	for _, p := range ps {
		if !managedMember(org.People, p) {
			report.PrintWarn(p.GetLogin() + " exists in github but not in manifest")
		} else {
			report.PrintInfo(p.GetLogin() + " exists in github")
		}

		report.Println()
	}

	err = inviteMembers(ctx, org.Name, missingMembers(org.People, ps), dry)
	if err != nil {
		return handleError(cmd, err)
	}

	return nil
}

func managedMember(manifestMembers []*gh_pb.People, member *github.User) bool {
	for _, mm := range manifestMembers {
		if strings.EqualFold(mm.Username, *member.Login) {
			return true
		}
	}

	return false
}

func missingMembers(manifestMembers []*gh_pb.People, githubMembers []*github.User) []*gh_pb.People {
	missing := []*gh_pb.People{}

	for _, mm := range manifestMembers {
		found := false
		for _, gm := range githubMembers {
			if strings.EqualFold(mm.Username, *gm.Login) {
				found = true
				break
			}
		}

		if !found {
			missing = append(missing, mm)
		}
	}

	return missing
}

func inviteMembers(ctx context.Context, org string, members []*gh_pb.People, dry bool) error {
	clt, err := client.ClientFromContext(ctx)
	if err != nil {
		return err
	}

	for _, m := range members {
		if dry {
			report.PrintAdd("invite " + m.Name)
			report.Println()
			continue
		}

		err := clt.InviteMember(ctx, org, m.Name)
		if err != nil {
			return err
		}

		report.PrintAdd("invited " + m.Name)
		report.Println()
	}

	return nil
}
