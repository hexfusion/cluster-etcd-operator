package member

import (
	"os"

	"github.com/spf13/cobra"
)

func NewMemberCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "member",
	}

	cmd.AddCommand(NewMemberAddCommand(os.Stderr))

	return cmd
}
