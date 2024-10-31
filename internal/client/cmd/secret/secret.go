package secret

import (
	"github.com/spf13/cobra"
)

var SecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "secret crud commands",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	SecretCmd.AddCommand(AddCmd)
	SecretCmd.AddCommand(ListCmd)
	SecretCmd.AddCommand(GetCmd)
	SecretCmd.AddCommand(DeleteCmd)
	SecretCmd.AddCommand(SyncCmd)
}
