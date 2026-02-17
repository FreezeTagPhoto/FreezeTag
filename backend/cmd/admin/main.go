package admin

import "github.com/spf13/cobra"

func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "update the password of a user",
		Run:   runAdminCommand,
	}
	return cmd
}

func runAdminCommand(cmd *cobra.Command, args []string) {
	
}